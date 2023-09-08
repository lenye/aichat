// Copyright 2023 The aichat Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chatgpt

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/lenye/aichat/internal/config"
	"github.com/lenye/aichat/pkg/web/logging"
	"github.com/lenye/aichat/pkg/web/sse"
)

// chatErr HttpChatCompletion 错误处理
func chatErr(tag string, err error, chStr chan<- string, logger *slog.Logger) error {

	var (
		reqErr *openai.RequestError
		apiErr *openai.APIError
		urlErr *url.Error
	)

	if errors.As(err, &reqErr) {
		logger.Error(tag,
			"error", reqErr,
			"type", "openai.RequestError",
		)
		chStr <- "[[错误请求]]"
		return nil
	} else if errors.As(err, &apiErr) {
		logger.Error(tag,
			"error", apiErr,
			"type", "openai.APIError",
		)
		switch apiErr.HTTPStatusCode {
		case 504, 500, 503:
			// retry
			chStr <- "[[服务不可用]]"
		case 429:
			// retry
			chStr <- "[[太多请求]]"
		case 401:
			// unauthorized
			chStr <- "[[未授权]]"
		case 400:
			// bad request
			// if strings.Contains(apiErr.Message, "content filtering") {
			//
			// }
			chStr <- "[[错误请求]]"
		default:
			// bad request
			chStr <- "[[错误请求]]"
		}
		return nil
	} else if errors.As(err, &urlErr) {
		logger.Error(tag,
			"error", urlErr,
			"type", "url.Error",
		)
		if urlErr.Timeout() {
			chStr <- "[[请求超时]]"
		}
		return nil
	}
	chStr <- fmt.Sprintf("[[%s]]", err.Error())
	return err
}

// HttpChatCompletion 聊天api
func HttpChatCompletion(r *http.Request,
	cfg *config.OpenAIConfig,
	req *openai.ChatCompletionRequest,
	chStr chan<- string) {
	ctx := r.Context()
	logger := logging.FromContext(ctx)
	logger.Debug("HttpChatCompletion",
		"openai.ChatCompletionRequest", req,
	)

	client, err := NewOpenAIClient(cfg.ApiKey, cfg.ApiType, cfg.ApiBaseUrl, cfg.Proxy)
	if err != nil {
		logger.Error("NewOpenAIClient failed",
			"error", err,
			"config", cfg,
		)
		chStr <- fmt.Sprintf("[[%s]]", err.Error())
		close(chStr)
		return
	}

	if req.Stream {
		streamReader, err := client.CreateChatCompletionStream(ctx, *req)
		if err != nil {
			if err := chatErr("CreateChatCompletionStream failed", err, chStr, logger); err != nil {
				logger.Error("CreateChatCompletionStream failed",
					"error", err.Error(),
				)
			}
			close(chStr)
			return
		}
		defer streamReader.Close()

		for {
			select {
			case <-ctx.Done():
				// client close
				close(chStr)
				return
			default:
			}

			resp, err := streamReader.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					// Stream finished
				} else {
					logger.Error("read stream failed",
						"error", err,
					)
					chStr <- fmt.Sprintf("[[%s]]", err.Error())
				}
				close(chStr)
				return
			}

			logger.Debug("stream",
				"response", resp,
			)

			for _, choice := range resp.Choices {
				if choice.Delta.Content != "" {
					chStr <- choice.Delta.Content
				}
			}
		}
	} else {
		resp, err := client.CreateChatCompletion(ctx, *req)
		if err != nil {
			if err := chatErr("CreateChatCompletion failed", err, chStr, logger); err != nil {
				logger.Error("CreateChatCompletion failed",
					"error", err,
				)
			}
			chStr <- fmt.Sprintf("[[%s]]", err.Error())
			close(chStr)
			return
		}
		chStr <- resp.Choices[0].Message.Content
		close(chStr)
		return
	}
}

// SSEServerChatResponseProcess http sse 处理请求结果
func SSEServerChatResponseProcess(r *http.Request,
	streamID string,
	chStr <-chan string) string {
	ctx := r.Context()
	var messages strings.Builder
	for {
		select {
		case <-ctx.Done():
			return messages.String()
		case str, ok := <-chStr:
			if !ok {
				sse.Default().Publish(streamID, &sse.Event{Data: []byte("<br><br>")})
				// 已被关闭
				return messages.String()
			}
			messages.WriteString(str)
			str = strings.Replace(str, "\r", "", -1)
			str = strings.Replace(str, "\n", "<br>", -1)
			sse.Default().Publish(streamID, &sse.Event{Data: []byte(str)})
		}
	}
}

// HttpChatResponseProcess http sse 处理请求结果
func HttpChatResponseProcess(w http.ResponseWriter, r *http.Request,
	chStr <-chan string) string {
	ctx := r.Context()
	logger := logging.FromContext(ctx)
	flusher, _ := w.(http.Flusher)
	var messages strings.Builder
	for {
		select {
		case <-ctx.Done():
			return messages.String()
		case str, ok := <-chStr:
			if !ok {
				// 已被关闭
				return messages.String()
			}
			messages.WriteString(str)
			_, err := fmt.Fprintf(w, "data: %s\n\n", strings.Replace(str, "\n", "\ndata: ", -1))
			if err != nil {
				logger.Error("write stream failed",
					"error", err,
				)
				return ""
			}
			flusher.Flush()
		}
	}
}

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

package ai

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

const (
	Timeout = 180 * time.Second
)

var defaultHTTPClient = &http.Client{
	Timeout:   Timeout,
	Transport: http.DefaultTransport,
}

// BuildClient 客户端
func BuildClient(apiKey, apiType, baseURL, proxy string) (*openai.Client, error) {
	if apiKey == "" {
		return nil, ErrInvalidAPIKey
	}
	if baseURL == "" {
		return nil, ErrInvalidBaseURL
	}

	var cfg openai.ClientConfig

	APIType := openai.APIType(apiType)

	switch APIType {
	case openai.APITypeOpenAI:
		cfg = openai.DefaultConfig(apiKey)
		cfg.BaseURL = baseURL
	case openai.APITypeAzure:
		cfg = openai.DefaultAzureConfig(apiKey, baseURL)
	default:
		return nil, ErrInvalidAPIType
	}

	cfg.HTTPClient = defaultHTTPClient

	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return nil, ErrInvalidProxy
		}
		cfg.HTTPClient.Transport.(*http.Transport).Proxy = http.ProxyURL(proxyUrl)
	}

	return openai.NewClientWithConfig(cfg), nil
}

func Chat(client *openai.Client,
	stream bool,
	user, model, prompt string,
	maxTokens, history int) {
	ctx := context.Background()
	req := &openai.ChatCompletionRequest{
		Temperature:      0.7,
		TopP:             1,
		N:                1,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
		MaxTokens:        maxTokens,
		User:             user,
		Model:            model,
	}

	var (
		sMsg  *openai.ChatCompletionMessage  // 系统提示语
		hMsgs []openai.ChatCompletionMessage // 聊天记录
	)
	// 系统提示语
	if prompt != "" {
		sMsg = &openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt, // "you are a helpful chatbot"
		}
	}

	// 会话
	fmt.Println("Conversation")
	fmt.Println("---------------------")
	fmt.Print("> ")

	// 用户输入
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		input := s.Text()
		if strings.TrimSpace(input) != "" {
			var msgs []openai.ChatCompletionMessage
			// 系统提示语
			if sMsg != nil {
				msgs = append(msgs, *sMsg)
			}
			// 聊天记录
			if history > 0 {
				if len(hMsgs)/2 > history {
					hMsgs = hMsgs[2:]
				}
				msgs = append(msgs, hMsgs...)
			}
			// 用户输入的提示语
			uMsg := openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: input,
			}
			msgs = append(msgs, uMsg)
			req.Messages = msgs

			if msg := chatCompletion(ctx, client, req, stream); msg != nil {
				// 保存聊天记录
				hMsgs = append(hMsgs, uMsg)
				hMsgs = append(hMsgs, *msg)
			}
		}
		fmt.Print("> ")
	}
}

func chatCompletion(ctx context.Context,
	client *openai.Client,
	req *openai.ChatCompletionRequest,
	stream bool) *openai.ChatCompletionMessage {
	if stream {
		chatStream, err := client.CreateChatCompletionStream(ctx, *req)
		if err != nil {
			fmt.Printf("CreateChatCompletionStream error: %v\n\n", err)
			return nil
		}
		defer chatStream.Close()

		var sb strings.Builder
		for {
			resp, err := chatStream.Recv()
			if err != nil {
				// Stream finished
				if errors.Is(err, io.EOF) {
					fmt.Print("\n\n")
					return &openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: sb.String(),
					}
				} else {
					fmt.Printf("stream error: %v\n\n", err)
					return nil
				}
			}

			// fmt.Printf("Stream resp: %v\n", resp)

			// // Stream FinishReason
			// if resp.Choices[0].FinishReason != "" {
			// 	fmt.Printf("stream finish reason: %s", resp.Choices[0].FinishReason)
			// }

			msgContent := resp.Choices[0].Delta.Content
			sb.WriteString(msgContent)
			fmt.Printf("%s", msgContent)
		}
	}

	resp, err := client.CreateChatCompletion(ctx, *req)
	if err != nil {
		fmt.Printf("CreateChatCompletion error: %v\n\n", err)
		return nil
	}
	// ai 回复
	fmt.Printf("%s\n\n", resp.Choices[0].Message.Content)

	return &resp.Choices[0].Message
}

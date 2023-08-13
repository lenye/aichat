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

package chat

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/sashabaranov/go-openai"

	"github.com/lenye/aichat/internal/chatgpt"
	"github.com/lenye/aichat/internal/config"
	"github.com/lenye/aichat/pkg/web/logging"
	"github.com/lenye/aichat/pkg/web/render"
	"github.com/lenye/aichat/pkg/web/sse"
	"github.com/lenye/aichat/pkg/web/templatemap"
)

// SseMessage sse server 回复聊天响应消息
func SseMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.FromContext(ctx)

	m := templatemap.FromContext(ctx)

	streamID := r.PostFormValue("stream_id")
	prompt := r.PostFormValue("prompt")
	if streamID != "" && prompt != "" {
		in := &chatgpt.Message{
			StreamID:  streamID,
			ID:        "",
			User:      "",
			Model:     "",
			Prompt:    prompt,
			System:    "",
			Stream:    false,
			History:   0,
			MaxTokens: 0,
		}

		in.Stream, _ = strconv.ParseBool(r.PostFormValue("stream"))
		in.Model = r.PostFormValue("model")
		if in.Model == "" {
			in.Model = openai.GPT3Dot5Turbo
		}
		in.System = r.PostFormValue("system")
		// if uHis, err := strconv.ParseUint(r.PostFormValue("history"), 10, 0); err != nil {
		// 	in.History = uint(uHis)
		// }
		if uu, err := strconv.ParseUint(r.PostFormValue("max_tokens"), 10, 0); err == nil {
			in.MaxTokens = uint(uu)
		}

		logger.Debug("input",
			slog.Any("data", in),
		)

		// user prompt
		inMsg := strings.Replace(prompt, "\r", "", -1)
		inMsg = strings.Replace(inMsg, "\n", "<br>", -1)
		sse.Default().Publish(in.StreamID, &sse.Event{
			Data: []byte("<p class=\"has-text-info\">" + inMsg + "</p>"),
		})

		var hisMsg []openai.ChatCompletionMessage // 聊天记录
		chatReq := chatgpt.MakeChatRequest(in, hisMsg)
		chStr := make(chan string)

		var wg sync.WaitGroup
		wg.Add(1)
		// ai chat
		go func() {
			defer wg.Done()
			chatgpt.HttpChatCompletion(r, config.Default().OpenAI, chatReq, chStr)
		}()
		messages := chatgpt.SSEServerChatResponseProcess(r, in.StreamID, chStr)
		logger.Debug("ai",
			slog.String("msg", messages),
		)

		wg.Wait()

		// todo 计算token，保存账户余额，保存聊天记录

		m["stream_id"] = in.StreamID
		m["model"] = in.Model
		m["stream"] = strconv.FormatBool(in.Stream)
		m["system"] = in.System
		m["history"] = strconv.FormatUint(uint64(in.History), 10)
	} else {
		m["stream_id"] = getStreamID(w, r)
		m["model"] = openai.GPT3Dot5Turbo
		m["stream"] = "true"
		m["system"] = ""
		m["history"] = "0"
	}

	render.Html(w, r, "chat_input.gohtml", m)
}

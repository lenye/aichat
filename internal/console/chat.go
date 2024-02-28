// Copyright 2023-2024 The aichat Authors. All rights reserved.
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

package console

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/lenye/aichat/internal/chatgpt"
)

const promptInput = "(Press 'q' to quit) > "

func Chat(client *openai.Client, in *chatgpt.Message) {
	ctx := context.Background()
	var hisMsg []openai.ChatCompletionMessage // 聊天记录
	// 会话
	fmt.Println("---------------------")
	if in.System != "" {
		fmt.Println(in.System)
	}
	fmt.Print(promptInput)

	// 用户输入
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		input := strings.TrimSpace(s.Text())
		if input != "" {
			if input == "q" {
				return
			}
			in.Prompt = input
			req := chatgpt.MakeChatRequest(in, hisMsg)
			if msg, err := chatCompletion(ctx, client, req); err == nil {
				// 保存聊天记录
				hisMsg = append(hisMsg, req.Messages[len(req.Messages)-1])
				hisMsg = append(hisMsg, *msg)
			}
		}
		fmt.Print(promptInput)
	}
}

func chatCompletion(ctx context.Context,
	client *openai.Client,
	req *openai.ChatCompletionRequest) (*openai.ChatCompletionMessage, error) {
	if req.Stream {
		chatStream, err := client.CreateChatCompletionStream(ctx, *req)
		if err != nil {
			fmt.Printf("CreateChatCompletionStream faild, cause: %s\n\n", err)
			return nil, err
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
					}, nil
				} else {
					fmt.Printf("\n\nstream read faild, cause: %s\n\n", err)
					return nil, err
				}
			}

			msgContent := resp.Choices[0].Delta.Content
			sb.WriteString(msgContent)
			fmt.Printf("%s", msgContent)
		}
	}

	resp, err := client.CreateChatCompletion(ctx, *req)
	if err != nil {
		fmt.Printf("CreateChatCompletion faild, cause: %s\n\n", err)
		return nil, err
	}
	// ai 回复
	fmt.Printf("%s\n\n", resp.Choices[0].Message.Content)

	return &resp.Choices[0].Message, nil
}

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

	"github.com/sashabaranov/go-openai"
)

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

	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return nil, ErrInvalidProxy
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		cfg.HTTPClient = &http.Client{
			Transport: transport,
		}
	}

	return openai.NewClientWithConfig(cfg), nil
}

func Chat(client *openai.Client, user, model, prompt string, stream bool) {
	ctx := context.Background()
	req := &openai.ChatCompletionRequest{
		Temperature:      0.7,
		TopP:             1,
		N:                1,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
		User:             user,
		Model:            model,
		Messages:         make([]openai.ChatCompletionMessage, 0),
	}

	// 系统提示语
	if prompt != "" {
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt, // "you are a helpful chatbot"
		})
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
			// 用户输入的提示语
			req.Messages = append(req.Messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: s.Text(),
			})

			if msg := chatCompletion(ctx, client, req, stream); msg != nil {
				req.Messages = append(req.Messages, *msg)
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
			fmt.Printf("ChatCompletionStream error: %v\n", err)
			return nil
		}
		defer chatStream.Close()

		var sb strings.Builder
		for {
			response, err := chatStream.Recv()
			if err != nil {
				// Stream finished
				if errors.Is(err, io.EOF) {
					fmt.Print("\n\n")
					return &openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: sb.String(),
					}
				} else {
					fmt.Printf("Stream error: %v\n", err)
					return nil
				}
			}

			// fmt.Printf("Stream response: %v\n", response)

			sb.WriteString(response.Choices[0].Delta.Content)
			fmt.Printf("%s", response.Choices[0].Delta.Content)
		}
	}

	resp, err := client.CreateChatCompletion(ctx, *req)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return nil
	}
	// ai 回复
	fmt.Printf("%s\n\n", resp.Choices[0].Message.Content)

	return &resp.Choices[0].Message
}

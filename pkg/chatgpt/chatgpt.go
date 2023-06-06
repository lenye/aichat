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
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

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
	req := openai.ChatCompletionRequest{
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
		// 用户输入的提示语
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: s.Text(),
		})

		resp, err := client.CreateChatCompletion(ctx, req)
		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			continue
		}

		// ai 回复
		fmt.Printf("%s\n\n", resp.Choices[0].Message.Content)
		req.Messages = append(req.Messages, resp.Choices[0].Message)

		fmt.Print("> ")
	}
}

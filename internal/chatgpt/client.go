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

package chatgpt

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

// NewOpenAIClient 客户端
func NewOpenAIClient(apiKey, apiType, baseURL, proxy string) (*openai.Client, error) {
	if apiKey == "" {
		return nil, errors.New("missed api key")
	}
	if baseURL != "" {
		if _, err := url.Parse(baseURL); err != nil {
			return nil, fmt.Errorf("invalid base url: %q, cause %s", baseURL, err)
		}
	}
	var cfg openai.ClientConfig

	APIType := openai.APIType(strings.ToUpper(apiType))

	switch APIType {
	case openai.APITypeOpenAI:
		cfg = openai.DefaultConfig(apiKey)
		if baseURL != "" {
			cfg.BaseURL = baseURL
		}
	case openai.APITypeAzure:
		if baseURL == "" {
			return nil, errors.New("missed base url")
		}
		cfg = openai.DefaultAzureConfig(apiKey, baseURL)
	default:
		return nil, fmt.Errorf("invalid api type: %q", apiType)
	}

	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy: %q, cause %s", proxy, err)
		}
		defaultHTTPClient.Transport.(*http.Transport).Proxy = http.ProxyURL(proxyUrl)
	}

	cfg.HTTPClient = defaultHTTPClient

	return openai.NewClientWithConfig(cfg), nil
}

// MakeChatRequest 生成请求消息, history=不含系统提示语的聊天记录
func MakeChatRequest(in *Message, history []openai.ChatCompletionMessage) *openai.ChatCompletionRequest {
	var (
		sysMsg  *openai.ChatCompletionMessage  // 系统提示语
		chatMsg []openai.ChatCompletionMessage // 当前请求对话的聊天内容
	)
	// 系统提示语
	if in.System != "" {
		sysMsg = &openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: in.System, // "you are a helpful chatbot"
		}
	}
	// 系统提示语
	if sysMsg != nil {
		chatMsg = append(chatMsg, *sysMsg)
	}
	// 聊天记录
	if in.History > 0 {
		if len(history)/2 > int(in.History) {
			history = history[2:]
		}
		chatMsg = append(chatMsg, history...)
	}
	// 用户输入的提示语
	uMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: in.Prompt,
	}
	chatMsg = append(chatMsg, uMsg)

	return &openai.ChatCompletionRequest{
		Temperature:      0.7,
		TopP:             1,
		N:                1,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
		MaxTokens:        int(in.MaxTokens),
		Stream:           in.Stream,
		User:             in.User,
		Model:            in.Model,
		Messages:         chatMsg,
	}
}

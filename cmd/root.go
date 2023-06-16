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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lenye/aichat/pkg/ai"
	"github.com/lenye/aichat/pkg/version"
)

var rootCmd = &cobra.Command{
	Use:   "aichat",
	Short: "ai chat",
	Long: `ai chat, chatGPT
    Open source: https://github.com/lenye/aichat`,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli, err := ai.BuildClient(apiKey, apiType, baseUrl, proxy)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}
		if len(args) == 1 {
			prompt = args[0]
		}
		ai.Chat(cli, stream, user, model, prompt, int(maxTokens), int(history))
	},
}

var (
	apiKey  string
	apiType string
	baseUrl string
	proxy   string

	maxTokens uint
	user      string
	model     string
	stream    bool
	prompt    string // 系统提示语
	history   uint   // 聊天记录条数
)

const (
	flagApiKey  = "api_key"  // secret key
	flagApiType = "api_type" // api 类型
	flagBaseUrl = "base_url" // api base url
	flagProxy   = "proxy"    // proxy

	flagMaxTokens = "max_tokens" // max_tokens
	flagUser      = "user"       // 用户标识
	flagModel     = "model"      // model
	flagStream    = "stream"     // stream
	flagHistory   = "history"    // 聊天记录条数
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate(`{{printf "%s" .Version}}`)
	rootCmd.Version = version.Print()

	rootCmd.Flags().StringVarP(&apiKey, flagApiKey, "k", "", "secret key (required)")
	_ = rootCmd.MarkFlagRequired(flagApiKey)

	rootCmd.Flags().StringVarP(&apiType, flagApiType, "t", "OPEN_AI", "api type: OPEN_AI, AZURE")
	rootCmd.Flags().StringVarP(&baseUrl, flagBaseUrl, "u", "https://api.openai.com/v1", "api base url")
	rootCmd.Flags().StringVar(&proxy, flagProxy, "", "proxy")

	rootCmd.Flags().UintVarP(&maxTokens, flagMaxTokens, "m", 0, "max_tokens")
	rootCmd.Flags().StringVar(&user, flagUser, "", "user")
	rootCmd.Flags().StringVar(&model, flagModel, "gpt-3.5-turbo", "model")
	rootCmd.Flags().BoolVar(&stream, flagStream, false, "is stream")
	rootCmd.Flags().UintVar(&history, flagHistory, 0, "chat history")
}

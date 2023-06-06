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

	"github.com/lenye/aichat/pkg/chatgpt"
	"github.com/lenye/aichat/pkg/version"
)

var rootCmd = &cobra.Command{
	Use:   "aichat",
	Short: "ai chat",
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli, err := chatgpt.BuildClient(ApiKey, ApiType, BaseUrl, Proxy)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}
		if len(args) == 1 {
			Prompt = args[0]
		}
		chatgpt.Chat(cli, "", Model, Prompt, Stream)
	},
}

var (
	ApiType string
	ApiKey  string
	BaseUrl string
	Proxy   string
	Model   string
	Prompt  string // 系统提示语
	Stream  bool
)

const (
	apiType = "api_type" // api 类型
	apiKey  = "api_key"  // secret key
	baseUrl = "base_url" // base url
	proxy   = "proxy"    // 代理
	model   = "model"    // model
	stream  = "stream"   // stream
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

	rootCmd.Flags().StringVarP(&ApiKey, apiKey, "k", "", "secret key (required)")
	_ = rootCmd.MarkFlagRequired(apiKey)

	rootCmd.Flags().StringVarP(&ApiType, apiType, "t", "OPEN_AI", "api type")
	rootCmd.Flags().StringVarP(&BaseUrl, baseUrl, "u", "https://api.openai.com/v1", "base url")
	rootCmd.Flags().StringVar(&Proxy, proxy, "", "proxy")
	rootCmd.Flags().StringVar(&Model, model, "gpt-3.5-turbo", "model")
	rootCmd.Flags().BoolVar(&Stream, stream, false, "is stream")
}

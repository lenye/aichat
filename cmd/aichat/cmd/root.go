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

package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"

	"github.com/lenye/aichat/assets"
	"github.com/lenye/aichat/internal/chatgpt"
	"github.com/lenye/aichat/internal/config"
	"github.com/lenye/aichat/internal/console"
	"github.com/lenye/aichat/internal/router"
	"github.com/lenye/aichat/pkg/project"
	"github.com/lenye/aichat/pkg/version"
	"github.com/lenye/aichat/pkg/web/render"
	"github.com/lenye/aichat/pkg/web/sse"
)

var (
	appPath string

	cfg *config.Configuration
)

var root = &cobra.Command{
	Use:   "aichat",
	Short: "AI Chat",
	Long: fmt.Sprintf(`AI Chat
  Source: %s`, version.OpenSource),
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	Args: cobra.NoArgs,
	Run:  rootRun,
}

// flagRunningMode 控制台模式
var flagRunningMode string

const (
	consoleMode = "console"
	webMode     = "web"
)

func Execute() {

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := root.Execute(); err != nil {
		logger := slog.Default()
		logger.Error("invalid command flags",
			"error", err,
		)
		// fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	if project.DevMode() {
		appPath = project.Root("cmd", "aichat")
	}
	cfg = config.New(appPath)

	root.SetVersionTemplate(`{{printf "%s" .Version}}`)
	root.Version = version.Print()

	// console 控制台模式
	root.Flags().StringVar(&flagRunningMode, "mode", "console", "running mode: console, web")

	// openai
	root.Flags().StringVar(&cfg.OpenAI.ApiType, "openai_api_type", string(openai.APITypeOpenAI), "openai api type: open_ai, azure")
	root.Flags().StringVar(&cfg.OpenAI.ApiKey, "openai_api_key", "", "openai api key (required)")
	_ = root.MarkFlagRequired("openai_api_key")
	root.Flags().StringVar(&cfg.OpenAI.ApiBaseUrl, "openai_api_base_url", "", "openai api base url")
	root.Flags().StringVar(&cfg.OpenAI.Proxy, "openai_proxy", "", "openai proxy")
	root.Flags().StringVar(&cfg.OpenAI.Model, "openai_model", openai.GPT3Dot5Turbo, "openai chat message model")
	root.Flags().StringVar(&cfg.OpenAI.System, "openai_system", "", "openai chat message system prompt")
	root.Flags().BoolVar(&cfg.OpenAI.SystemRaw, "openai_system_raw", false, "openai chat message system prompt without any escape processing")
	root.Flags().BoolVar(&cfg.OpenAI.Stream, "openai_stream", true, "openai chat message stream mode")
	root.Flags().UintVar(&cfg.OpenAI.MaxTokens, "openai_max_tokens", 0, "openai chat message max tokens")
	root.Flags().UintVar(&cfg.OpenAI.History, "openai_history", 0, "openai chat message history")

	// web server 在console模式下不用
	root.Flags().UintVar(&cfg.Web.Port, "web_port", 8080, "web server listen port")
	// web log 在console模式下不用
	root.Flags().StringVar(&cfg.Log.Level, "log_level", "info", "log message level: debug, info, warn, error")
	root.Flags().StringVar(&cfg.Log.Format, "log_format", "text", "log message encode format: text, json")
}

func rootRun(cmd *cobra.Command, args []string) {
	switch strings.ToLower(flagRunningMode) {
	case consoleMode, webMode:
	default:
		fmt.Println(fmt.Sprintf("invalid running mode: %q, use the default: %q", flagRunningMode, consoleMode))
		flagRunningMode = consoleMode
	}

	logger := slog.Default()
	if err := config.Setup(cfg); err != nil {
		logger = slog.Default()
		logger.Error("config setup failed",
			"error", err,
		)
		return
	}

	if project.DevMode() {
		cfg.Print()
	}

	if !cfg.OpenAI.SystemRaw {
		var err error
		cfg.OpenAI.System, err = project.StrRaw2Interpreted(cfg.OpenAI.System)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	if strings.ToLower(flagRunningMode) == consoleMode {
		cli, err := chatgpt.NewOpenAIClient(cfg.OpenAI.ApiKey, cfg.OpenAI.ApiType, cfg.OpenAI.ApiBaseUrl, cfg.OpenAI.Proxy)
		if err != nil {
			fmt.Println(err)
			return
		}
		in := &chatgpt.Message{
			Model:     cfg.OpenAI.Model,
			System:    cfg.OpenAI.System,
			Stream:    cfg.OpenAI.Stream,
			History:   cfg.OpenAI.History,
			MaxTokens: cfg.OpenAI.MaxTokens,
		}
		console.Chat(cli, in)
	} else {
		// html 模板
		render.SetDebug(project.DevMode())
		render.SetFileSystem(assets.HtmlFS())
		if err := render.LoadTemplates(); err != nil {
			logger.Error("render.LoadTemplates failed",
				"error", err,
			)
			return
		}

		wg := new(sync.WaitGroup)

		sseServer := sse.Default()
		sseServer.AutoStream = true
		sseServer.AutoReplay = false
		sse.SetDefault(sseServer)
		defer sseServer.Close()

		httpd, err := config.WebListenAndServe(router.New(), cfg.Web, wg, logger)
		if err != nil {
			return
		}

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(signalChan)

		s := <-signalChan
		logger.Debug("received signal",
			"signal", s,
		)

		config.WebShutdown(httpd, logger)

		wg.Wait()
	}
}

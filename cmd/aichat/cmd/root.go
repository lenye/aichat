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
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"

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

	cfg *config.Config
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

// flagConsole 控制台模式
var flagConsole bool

func Execute() {

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := root.Execute(); err != nil {
		logger := slog.Default()
		logger.Error("invalid command flags",
			slog.Any("error", err),
		)
		// fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	if project.DevMode() {
		appPath = project.Root("cmd", "app")
	}
	cfg = config.New(appPath)

	root.SetVersionTemplate(`{{printf "%s" .Version}}`)
	root.Version = version.Print()

	// console 控制台模式
	root.Flags().BoolVar(&flagConsole, "console", false, "running console mode")

	// openai
	root.Flags().StringVar(&cfg.OpenAI.ApiType, "openai_api_type", string(openai.APITypeOpenAI), "openai api type: OPEN_AI, AZURE")
	root.Flags().StringVar(&cfg.OpenAI.ApiKey, "openai_api_key", "", "openai api key (required)")
	_ = root.MarkFlagRequired("openai_api_key")
	root.Flags().StringVar(&cfg.OpenAI.ApiBaseUrl, "openai_api_base_url", "", "openai api base url")
	root.Flags().StringVar(&cfg.OpenAI.Proxy, "openai_proxy", "", "openai proxy")
	root.Flags().StringVar(&cfg.OpenAI.Model, "openai_model", openai.GPT3Dot5Turbo, "openai chat message model")
	root.Flags().StringVar(&cfg.OpenAI.System, "openai_system", "", "openai chat message role system")
	root.Flags().BoolVar(&cfg.OpenAI.Stream, "openai_stream", true, "openai chat message stream mode")
	root.Flags().UintVar(&cfg.OpenAI.MaxTokens, "openai_max_tokens", 0, "openai chat message max tokens")
	root.Flags().UintVar(&cfg.OpenAI.History, "openai_history", 0, "openai chat message history")

	// log 在console模式下不用
	root.Flags().BoolVar(&cfg.Log.Caller, "log_caller", false, "log annotate each message with the filename, line number and function name")
	root.Flags().StringVar(&cfg.Log.Level, "log_level", "INFO", "log message level: DEBUG, INFO, WARN, ERROR")
	root.Flags().StringVar(&cfg.Log.Format, "log_format", "TEXT", "log message encode format: TEXT, JSON")

	// http server 在console模式下不用
	root.Flags().UintVar(&cfg.HttpServer.Port, "httpd_port", 8080, "httpd listen port")
}

func rootRun(cmd *cobra.Command, args []string) {
	var start = time.Now()

	logger := slog.Default()
	if err := config.Setup(cfg); err != nil {
		logger = slog.Default()
		logger.Error("config setup failed",
			slog.Any("error", err),
		)
		// fmt.Println(err)
		return
	}

	if project.DevMode() {
		cfg.Print()
	}

	if flagConsole {
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
				slog.Any("error", err),
			)
			return
		}

		wg := new(sync.WaitGroup)

		sseServer := sse.Default()
		sseServer.AutoStream = true
		sseServer.AutoReplay = false
		sse.SetDefault(sseServer)
		defer sseServer.Close()

		httpd, err := config.HttpListenAndServe(router.New(), cfg.HttpServer, wg, logger)
		if err != nil {
			return
		}

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(signalChan)

		s := <-signalChan
		logger.Debug("received signal",
			slog.Any("signal", s),
		)

		config.HttpShutdown(httpd, logger)

		wg.Wait()
	}

	logger.Info(version.AppName+" exit",
		slog.Duration("uptime", time.Since(start)),
	)
}

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

package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/sashabaranov/go-openai"

	"github.com/lenye/aichat/pkg/project"
)

var defaultConfig atomic.Value

func init() {
	defaultConfig.Store(New(""))
}

// Default returns the default Configuration.
func Default() *Configuration {
	return defaultConfig.Load().(*Configuration)
}

// SetDefault makes v the default Configuration.
func SetDefault(v *Configuration) {
	defaultConfig.Store(v)
}

// New init config
func New(appDirIn string) *Configuration {
	v := &Configuration{
		App: &AppConfig{
			Path: project.ExecAppPath(),
		},
		Log: &LogConfig{
			Caller: false,
			Level:  "info",
			Format: "text",
		},
		Web:    new(WebServerConfig),
		OpenAI: new(OpenAIConfig),
	}
	if appDirIn == "" {
		v.App.Dir = filepath.Dir(v.App.Path)
	} else {
		v.App.Dir = appDirIn
	}

	return v
}

// Configuration 配置
type Configuration struct {
	App    *AppConfig       `json:"app"`    // 程序运行目录
	Log    *LogConfig       `json:"log"`    // 日志
	Web    *WebServerConfig `json:"web"`    // web server
	OpenAI *OpenAIConfig    `json:"openai"` // openai
}

// Print 打印配置
func (p *Configuration) Print() {
	slog.Debug("configuration",
		slog.Group("config",
			"app", p.App, "log", p.Log, "web", p.Web, "openai", p.OpenAI,
		),
	)
}

// AppConfig app config
type AppConfig struct {
	Path string `json:"path"` // C:\opt\_test\option.test.exe
	Dir  string `json:"dir"`  // C:\opt\_test
}

// LogConfig 日志配置
type LogConfig struct {
	Caller bool   `yaml:"caller,omitempty"` // true=打印代码名称和行号
	Level  string `yaml:"level,omitempty"`  // 输出日志level
	Format string `yaml:"format,omitempty"` // 日志输出格式 text, json
}

// WebServerConfig web server配置
type WebServerConfig struct {
	Port uint `json:"port"` // 服务端口
}

// OpenAIConfig chatGPT配置
type OpenAIConfig struct {
	ApiType    string `json:"api_type,omitempty"`     // api类型
	ApiKey     string `json:"api_key"`                // api key
	ApiBaseUrl string `json:"api_base_url,omitempty"` // base url
	Proxy      string `json:"proxy,omitempty"`        // 代理
	Model      string `json:"model"`                  // 运行模式
	System     string `json:"system,omitempty"`       // 系统提示信息
	SystemRaw  bool   `json:"system_raw,omitempty"`   // 系统提示信息不支持“\”转义
	Stream     bool   `json:"stream"`                 // 流模式
	MaxTokens  uint   `json:"max_tokens"`             // 最大tokens
	History    uint   `json:"history"`                // 历史记录
}

func setupLog(v *LogConfig) {
	opts := &slog.HandlerOptions{
		AddSource: false,
	}
	var (
		err error
		lvv slog.LevelVar
	)
	lv := strings.ToUpper(v.Level)
	switch lv {
	case slog.LevelDebug.String():
		lvv.Set(slog.LevelDebug)
		opts.AddSource = true
	case slog.LevelInfo.String():
		lvv.Set(slog.LevelInfo)
	case slog.LevelWarn.String():
		lvv.Set(slog.LevelWarn)
	case slog.LevelError.String():
		lvv.Set(slog.LevelError)
	default:
		lvv.Set(slog.LevelInfo)
		err = fmt.Errorf("invalid log_level: %q, use the default level: %q", v.Level, slog.LevelInfo.String())
	}
	opts.Level = lvv.Level()

	var handler slog.Handler
	ft := strings.ToUpper(v.Format)
	switch ft {
	case "TEXT":
		handler = slog.NewTextHandler(os.Stdout, opts)
	case "JSON":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("invalid log_format: %q, use the default format: %q", v.Format, "TEXT"),
			)
		}
	}
	logger := slog.New(handler)
	if err != nil {
		logger.Warn("user the default log flags",
			"error", err,
		)
	}
	slog.SetDefault(logger)
}

func checkOpenAIConfig(v *OpenAIConfig) error {
	// chat
	if v.ApiType == "" {
		v.ApiType = string(openai.APITypeOpenAI)
	} else {
		apiType := openai.APIType(strings.ToUpper(v.ApiType))
		switch apiType {
		case openai.APITypeOpenAI, openai.APITypeAzure, openai.APITypeAzureAD:
		default:
			return fmt.Errorf("invalid openai_api_type: %q", v.ApiType)
		}
	}

	if v.ApiBaseUrl != "" {
		if _, err := url.Parse(v.ApiBaseUrl); err != nil {
			return fmt.Errorf("invalid openai_api_base_url: %q, cause: %w", v.ApiBaseUrl, err)
		}
	}

	if v.Proxy != "" {
		if _, err := url.Parse(v.Proxy); err != nil {
			return fmt.Errorf("invalid openai_proxy: %q, cause: %w", v.Proxy, err)
		}
	}
	return nil
}

func Setup(v *Configuration) error {
	// log
	setupLog(v.Log)

	// openai
	if err := checkOpenAIConfig(v.OpenAI); err != nil {
		return err
	}

	SetDefault(v)

	return nil
}

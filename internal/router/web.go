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

package router

import (
	"net/http"

	"github.com/lenye/aichat/assets"
	"github.com/lenye/aichat/internal/handler/chat"
	"github.com/lenye/aichat/pkg/project"
	"github.com/lenye/aichat/pkg/web/alice"
	"github.com/lenye/aichat/pkg/web/middleware"
	"github.com/lenye/aichat/pkg/web/sse"
)

func New() http.Handler {
	r := http.NewServeMux()
	name := "web"
	stdPipe := alice.New(
		middleware.AccessLog(name),
	)

	// static
	staticFS := assets.StaticFS()
	fileServer := http.FileServer(http.FS(staticFS))
	staticChain := alice.New(middleware.ConfigureStaticAssets(project.DevMode()))
	r.Handle("GET /favicon.ico", staticChain.Then(fileServer))
	r.Handle("GET /static/", staticChain.Then(http.StripPrefix("/static/", fileServer)))

	// sse
	r.Handle("GET /chat/sse", stdPipe.ThenFunc(sse.Default().ServeHTTP))

	// tpl
	tplPipe := stdPipe.Append(middleware.TemplateMap)
	r.Handle("GET /chat", tplPipe.ThenFunc(chat.Chat))
	r.Handle("POST /chat/sse/msg", tplPipe.ThenFunc(chat.SseMessage))
	r.Handle("POST /chat/msg", tplPipe.ThenFunc(chat.Message))

	return r
}

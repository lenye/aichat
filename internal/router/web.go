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

package router

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/lenye/aichat/assets"
	"github.com/lenye/aichat/internal/handler"
	"github.com/lenye/aichat/internal/handler/chat"
	"github.com/lenye/aichat/pkg/project"
	"github.com/lenye/aichat/pkg/web/alice"
	"github.com/lenye/aichat/pkg/web/middleware"
	"github.com/lenye/aichat/pkg/web/sse"
)

func New() *httprouter.Router {
	r := httprouter.New()
	name := "httpd"
	stdPipe := alice.New(
		middleware.AccessLog(name),
	)

	r.PanicHandler = handler.Panic()
	r.NotFound = stdPipe.ThenFunc(handler.NotFound)

	r.HandleOPTIONS = true
	// set a global OPTIONS handler
	r.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Adjust status code to 204
		w.WriteHeader(http.StatusNoContent)
	})
	r.HandleMethodNotAllowed = true
	r.MethodNotAllowed = stdPipe.ThenFunc(handler.MethodNotAllowed)

	// static
	staticFS := assets.StaticFS()
	fileServer := http.FileServer(http.FS(staticFS))
	staticChain := alice.New(middleware.ConfigureStaticAssets(project.DevMode()))
	r.Handler(http.MethodGet, "/favicon.ico", staticChain.Then(fileServer))
	r.Handler(http.MethodGet, "/static/*filepath", staticChain.Then(http.StripPrefix("/static/", fileServer)))

	// sse
	r.Handler(http.MethodGet, "/chat/sse", stdPipe.ThenFunc(sse.Default().ServeHTTP))

	// tpl
	tplPipe := stdPipe.Append(middleware.TemplateMap)
	r.Handler(http.MethodGet, "/chat", tplPipe.ThenFunc(chat.Chat))
	r.Handler(http.MethodPost, "/chat/sse/msg", tplPipe.ThenFunc(chat.SseMessage))
	// r.Handler(http.MethodPost, "/chat/msg", tplPipe.ThenFunc(chat.Message))

	return r
}

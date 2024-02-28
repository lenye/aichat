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

package handler

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/lenye/aichat/pkg/web"
	"github.com/lenye/aichat/pkg/web/realip"
	"github.com/lenye/aichat/pkg/web/render"
)

// NotFound replies to the request with an HTTP 404 not found error.
func NotFound(w http.ResponseWriter, r *http.Request) {
	code := http.StatusNotFound
	render.HtmlStatus(w, r, code, "404.gohtml", map[string]string{
		"title": http.StatusText(code),
		"error": http.StatusText(code)})
}

// MethodNotAllowed replies to the request with an HTTP 405 method not allowed error.
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	code := http.StatusMethodNotAllowed
	render.HtmlStatus(w, r, code, "405.gohtml", map[string]string{
		"title": http.StatusText(code),
		"error": http.StatusText(code)})
}

// InternalServerError replies to the request with an HTTP 500 method not allowed error.
func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	render.Html500(w, r, err)
}

// Panic 异常
func Panic() func(w http.ResponseWriter, r *http.Request, p any) {
	return func(w http.ResponseWriter, r *http.Request, p any) {
		start := time.Now()
		ww := web.NewResponseWriterWrapper(w)

		InternalServerError(ww, r, p.(error))

		slog.Error("panic",
			"duration", time.Since(start),
			"status", ww.StatusCode,
			"method", r.Method,
			"url", r.URL,
			"ip", realip.ClientIP(r),
			"user_agent", r.UserAgent(),
			"error", p,
			"Stack", string(debug.Stack()),
		)
	}
}

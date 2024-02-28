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

package chat

import (
	"net/http"
	"strconv"

	"github.com/lenye/aichat/internal/config"
	"github.com/lenye/aichat/pkg/requestid"
	"github.com/lenye/aichat/pkg/web/render"
	"github.com/lenye/aichat/pkg/web/templatemap"
)

const (
	cookieName = "stream_id"
)

func Chat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	m := templatemap.FromContext(ctx)

	cfg := config.Default()

	m["stream_id"] = getStreamID(w, r)
	m["model"] = cfg.OpenAI.Model
	m["stream"] = strconv.FormatBool(cfg.OpenAI.Stream)
	m["system"] = cfg.OpenAI.System
	m["max_tokens"] = strconv.Itoa(int(cfg.OpenAI.MaxTokens))
	m["history"] = strconv.Itoa(int(cfg.OpenAI.History))

	render.Html(w, r, "chat.gohtml", m)
}

func getStreamID(w http.ResponseWriter, r *http.Request) string {
	var (
		cookie *http.Cookie
		err    error
	)

	// Chrome 会将到期日期限制为允许的最大值：自设置 Cookie 之日起 400 天
	cookie, err = r.Cookie(cookieName)
	if err != nil {
		cookie = &http.Cookie{
			Name:   cookieName,
			Value:  requestid.New(),
			MaxAge: 86400 * 400, // 86400 = 24 hours in seconds
		}
		http.SetCookie(w, cookie)
	}
	if len(cookie.Value) != 32 {
		cookie = &http.Cookie{
			Name:   cookieName,
			Value:  requestid.New(),
			MaxAge: 86400 * 400, // 86400 = 24 hours in seconds
		}
		http.SetCookie(w, cookie)
	}
	return cookie.Value
}

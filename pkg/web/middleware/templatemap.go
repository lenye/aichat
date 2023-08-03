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

package middleware

import (
	"net/http"

	"github.com/lenye/aichat/pkg/project"
	"github.com/lenye/aichat/pkg/version"
	"github.com/lenye/aichat/pkg/web/templatemap"
)

// TemplateMap handler
func TemplateMap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		m := templatemap.FromContext(ctx)
		m["title"] = version.AppName
		m["buildTime"] = version.BuildTime
		m["buildCommit"] = version.BuildCommit
		m["source"] = version.OpenSource
		m["version"] = version.Version

		m["mode"] = project.DevMode()
		m["requestUrl"] = r.URL.String()

		ctx = templatemap.WithContext(ctx, m)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

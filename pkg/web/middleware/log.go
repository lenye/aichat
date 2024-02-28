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

package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/lenye/aichat/pkg/web"
	"github.com/lenye/aichat/pkg/web/logging"
	"github.com/lenye/aichat/pkg/web/realip"
)

// AccessLog 访问日志
func AccessLog(name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := web.NewResponseWriterWrapper(w)
			ctx := r.Context()

			// logger
			logger := slog.Default().WithGroup(name)

			ctx = logging.WithContext(ctx, logger)
			next.ServeHTTP(ww, r.WithContext(ctx))

			logger.Info("access",
				"duration", time.Since(start),
				"status", ww.StatusCode,
				"method", r.Method,
				"url", r.URL,
				"size", ww.ContentLength,
				"ip", realip.ClientIP(r),
				"user_agent", r.UserAgent(),
			)
		})
	}
}

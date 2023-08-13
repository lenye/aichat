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

package render

import (
	"bytes"
	"fmt"
	"html"
	"log/slog"
	"net/http"

	"github.com/lenye/aichat/pkg/web/logging"
)

const HdrHtml = "text/html"

// HtmlNoContent sets 204 status code
func HtmlNoContent(w http.ResponseWriter) {
	writeNoCacheResponseContentType(w, HdrHtml)
	w.WriteHeader(http.StatusOK)
}

// Html calls HtmlStatus with a http.StatusOK (200).
func Html(w http.ResponseWriter, r *http.Request, tmpl string, data any) {
	HtmlStatus(w, r, http.StatusOK, tmpl, data)
}

// HtmlStatus renders the given Html template by name. It attempts to
// gracefully handle any rendering errors to avoid partial responses sent to the
// response by writing to a buffer first, then flushing the buffer to the
// response.
//
// If template rendering fails, a generic 500 page is returned. In dev mode, the
// error is included on the page. If flushing the buffer to the response fails,
// an error is logged, but no recovery is attempted.
//
// The buffers are fetched via a sync.Pool to reduce allocations and improve
// performance.
func HtmlStatus(w http.ResponseWriter, r *http.Request, statusCode int, tmpl string, data any) {
	writeResponseContentType(w, HdrHtml)
	ctx := r.Context()
	// Hello there reader! If you've made it here, you're likely wondering why
	// you're getting an error about response codes. For client-interop, it's very
	// important that we retain and maintain the allowed list of response codes.
	// Adding a new response statusCode requires coordination with the client team so
	// they can update their applications to handle that new response statusCode.
	if !AllowedResponseCode(statusCode) {
		logging.FromContext(ctx).Error("unregistered response statusCode",
			slog.Int("statusCode", statusCode),
			slog.String("func", "HtmlStatus"),
		)

		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf("%d is not a registered response statusCode", statusCode)
		if _, wErr := fmt.Fprintf(w, htmlErrTmpl, msg); wErr != nil {
			logging.FromContext(ctx).Error("failed to write html to response",
				slog.Any("error", wErr),
				slog.String("func", "HtmlStatus"),
			)
		}
		return
	}

	if isDebug {
		if err := LoadTemplates(); err != nil {
			logging.FromContext(ctx).Error("failed to reload templates in renderer",
				slog.Any("error", err),
				slog.String("func", "HtmlStatus"),
			)

			msg := html.EscapeString(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			if _, wErr := fmt.Fprintf(w, htmlErrTmpl, msg); wErr != nil {
				logging.FromContext(ctx).Error("failed to write html to response",
					slog.Any("error", wErr),
					slog.String("func", "HtmlStatus"),
				)
			}
			return
		}
	}

	// Acquire a renderer
	b := rendererPool.Get().(*bytes.Buffer)
	b.Reset()
	defer rendererPool.Put(b)

	// Render into the renderer
	if err := executeHTMLTemplate(b, tmpl, data); err != nil {
		logging.FromContext(ctx).Error("failed to execute html template",
			slog.Any("error", err),
			slog.String("func", "HtmlStatus"),
		)

		msg := "An internal error occurred."
		if isDebug {
			msg = err.Error()
		}
		msg = html.EscapeString(msg)

		w.WriteHeader(http.StatusInternalServerError)
		if _, wErr := fmt.Fprintf(w, htmlErrTmpl, msg); wErr != nil {
			logging.FromContext(ctx).Error("failed to write html to response",
				slog.Any("error", wErr),
				slog.String("func", "HtmlStatus"),
			)
		}
		return
	}

	// Rendering worked, flush to the response
	w.WriteHeader(statusCode)
	if _, err := b.WriteTo(w); err != nil {
		// We couldn't write the buffer. We can't change the response header or
		// content type if we got this far, so the best option we have is to log the
		// error.
		logging.FromContext(ctx).Error("failed to write html to response",
			slog.Any("error", err),
			slog.String("func", "HtmlStatus"),
		)
	}
}

// Html500 renders the given error as Html. In production mode, this always
// renders a generic "server error" message. In isDebug, it returns the actual
// error from the caller.
func Html500(w http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	errStr := http.StatusText(code)

	if isDebug {
		errStr = err.Error()
	}
	HtmlStatus(w, r, code, "500.gohtml", map[string]string{
		"title": http.StatusText(code),
		"error": errStr,
	})
}

// htmlErrTmpl is the template to use when returning an Html error. It is
// rendered using Printf, not html/template, so values must be escaped by the
// caller.
const htmlErrTmpl = `
<html>
  <head>
    <title>Internal server error</title>
  </head>
  <body>
    <h1>Internal server error</h1>
    <p style="font-family:monospace">%s</p>
  </body>
</html>
`

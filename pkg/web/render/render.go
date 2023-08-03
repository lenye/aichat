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
	"errors"
	"fmt"
	htmltemplate "html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"sync"
)

var (
	// rendererPool is a pool of *bytes.Buffer, used as a rendering buffer to
	// prevent partial responses being sent to clients.
	rendererPool = &sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 0, 1024))
		},
	}

	// templates is the actually collection of templates. templatesLoader is a
	// function for (re)loading templates. templatesLock is a mutex to prevent
	// concurrent modification of the templates field.
	templates     *htmltemplate.Template
	templatesLock sync.RWMutex

	fileSystem fs.FS

	isDebug bool
)

func SetDebug(v bool) {
	isDebug = v
}

// writeResponseContentType 响应内容类型
func writeResponseContentType(w http.ResponseWriter, value string) {
	w.Header().Set("Content-Type", value)
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

// writeNoCacheResponseContentType 响应内容类型
func writeNoCacheResponseContentType(w http.ResponseWriter, value string) {
	writeResponseContentType(w, value)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
}

// allowedResponseCodes are the list of allowed response codes. This is
// primarily here to catch if someone, in the future, accidentally includes a
// bad status code.
var allowedResponseCodes = map[int]struct{}{
	http.StatusOK:                    {},
	http.StatusBadRequest:            {},
	http.StatusUnauthorized:          {},
	http.StatusNotFound:              {},
	http.StatusMethodNotAllowed:      {},
	http.StatusConflict:              {},
	http.StatusPreconditionFailed:    {},
	http.StatusRequestEntityTooLarge: {},
	http.StatusTooManyRequests:       {},
	http.StatusInternalServerError:   {},
}

// AllowedResponseCode returns true if the code is a permitted response code,
// false otherwise.
func AllowedResponseCode(code int) bool {
	_, ok := allowedResponseCodes[code]
	return ok
}

// executeHTMLTemplate executes a single Html template with the provided data.
func executeHTMLTemplate(w io.Writer, name string, data any) error {
	templatesLock.RLock()
	defer templatesLock.RUnlock()

	if templates == nil {
		return errors.New("no html templates are defined")
	}

	return templates.ExecuteTemplate(w, name, data)
}

func SetFileSystem(fsys fs.FS) {
	fileSystem = fsys
}

// LoadTemplates loads or reloads all templates.
func LoadTemplates() error {
	templatesLock.Lock()
	defer templatesLock.Unlock()

	if fileSystem == nil {
		return nil
	}

	htmltpl := htmltemplate.New("").
		Option("missingkey=zero")

	if err := loadTemplates(fileSystem, htmltpl); err != nil {
		return fmt.Errorf("load templates failed, cause: %w", err)
	}

	templates = htmltpl
	return nil
}

func loadTemplates(fsys fs.FS, htmltmpl *htmltemplate.Template) error {
	// You might be thinking to yourself, wait, why don't you just use
	// template.ParseFS(fsys, "**/*.html"). Well, still as of Go 1.16, glob
	// doesn't support shopt globbing, so you still have to walk the entire
	// filepath.
	return fs.WalkDir(fsys, ".", func(pth string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".gohtml") {
			if _, err := htmltmpl.ParseFS(fsys, pth); err != nil {
				return fmt.Errorf("failed to parse %q, cause: %w", pth, err)
			}
		}

		return nil
	})
}

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

package assets

import (
	"embed"
	"io/fs"
	"os"

	"github.com/lenye/aichat/pkg/project"
)

//go:embed html html/**/*
var _htmlFS embed.FS

// This gets around an inconsistency where the embed is rooted at server/, but
// the os.DirFS is rooted after server/.
var htmlFS, _ = fs.Sub(_htmlFS, "html")

// HtmlFS returns the file system for the server  assets.
func HtmlFS() fs.FS {
	if project.DevMode() {
		return os.DirFS(project.Root("assets", "html"))
	}

	return htmlFS
}

var staticFS, _ = fs.Sub(htmlFS, "static")

// StaticFS returns the file system for the server static assets, rooted
// at static/.
func StaticFS() fs.FS {
	if project.DevMode() {
		return os.DirFS(project.Root("assets", "html", "static"))
	}

	return staticFS
}

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

package templatemap

import "fmt"

// TemplateMap is a typemap for the HTML templates.
type TemplateMap map[string]any

// Title sets the title on the template map. If a title already exists, the new
// value is prepended.
func (m TemplateMap) Title(f string, args ...any) {
	if f == "" {
		return
	}

	s := f
	if len(args) > 0 {
		s = fmt.Sprintf(f, args...)
	}

	if current := m["title"]; current != nil && current != "" {
		m["title"] = fmt.Sprintf("%s | %s", s, current)
		return
	}

	m["title"] = s
}

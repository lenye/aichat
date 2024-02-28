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

import (
	"context"

	"github.com/lenye/aichat/pkg/web/contextkey"
)

// templateMap context key
var tplMapCtxKey = contextkey.New("template.map")

func WithContext(ctx context.Context, v TemplateMap) context.Context {
	return context.WithValue(ctx, tplMapCtxKey, v)
}

func FromContext(ctx context.Context) TemplateMap {
	v := ctx.Value(tplMapCtxKey)
	if v == nil {
		return make(TemplateMap)
	}

	m, ok := v.(TemplateMap)
	if !ok {
		return make(TemplateMap)
	}

	return m
}

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

package sse

import (
	"context"

	"github.com/lenye/aichat/pkg/web/contextkey"
)

var sseCtxKey = contextkey.New("sse")

func WithContext(ctx context.Context, v *Server) context.Context {
	return context.WithValue(ctx, sseCtxKey, v)
}

func FromContext(ctx context.Context) *Server {
	v := ctx.Value(sseCtxKey)
	if v == nil {
		return nil
	}

	vv, ok := v.(*Server)
	if !ok {
		return nil
	}
	return vv
}

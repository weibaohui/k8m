package response

import (
	"context"
	"net/http"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := New(w, r)
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxKey, c)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetContext(r *http.Request) *Context {
	if ctx, ok := r.Context().Value(ctxKey).(*Context); ok {
		return ctx
	}
	return nil
}

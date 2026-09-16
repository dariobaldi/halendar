package main

import (
	"context"
	"net/http"

	"github.com/dariobaldi/halendar_back/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *app) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *app) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}

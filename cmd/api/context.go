package main

import (
	"context"
	"net/http"

	"github.com/ildx/greenlight/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

// add user to request context
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// get user from request context
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}

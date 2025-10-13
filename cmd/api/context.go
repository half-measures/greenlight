package main

//Idea is to store data during arbitray data during the lifetime of the request
//this is helper methods to read/write user struct to and from req context
import (
	"context"
	"net/http"

	"greenlight.alexedwards.net/internal/data"
)

type contextKey string

// Voncert string user to above type and assign it
// obj is to getting and setting user info in req context
const userContextKey = contextKey("user")

// add struct to the context
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// the context set user gets user struct req context
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}

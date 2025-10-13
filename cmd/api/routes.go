package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	//init new router instance
	router := httprouter.New()

	//custom errors.go for 404 handling
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	//custom errors.go for 405 handling
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	//register 'normal' methods
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	//Route below for POST users endpoint to create a user
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	//Put rather than post for idempotenty
	//
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	//return routerhttp instance
	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}

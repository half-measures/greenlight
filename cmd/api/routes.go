package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	//init new router intance
	router := httprouter.New()

	//custom errors.go for 404 handling
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	//custom errors.go for 405 handling
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	//register methods
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	//return routerhttp instance
	return router
}

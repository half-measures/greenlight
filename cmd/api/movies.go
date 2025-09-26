package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// createmovie handler for POST /v1/movies endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie") //temp response
}

// GET showmoviehandler endpoint, for now gets id param from current URL and give response
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// when httprouter is parsing request, url is stored in request context
	// lets get it front and center
	params := httprouter.ParamsFromContext(r.Context())

	// use byname to get value of id param from above slice.
	// if id is invalid, return 404
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	// otherwise, show movie details for now
	fmt.Fprintf(w, "show the details of the movie %d\n", id)
}

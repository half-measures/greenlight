package main

import (
	"fmt"
	"net/http"
	"time"

	"greenlight.alexedwards.net/internal/data"
)

// createmovie handler for POST /v1/movies endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	//declare anon struct to hold info we expect to be in the http body
	//Struct will be the *target decode destination
	var input struct {
		Title   string        `json:"title"`
		Year    int32         `json:"year"`
		Runtime data.Runetime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}
	//init new decoder to read from request body and put into input struct
	//if error during decoder send 400 reponse using our custom errs
	//Decode must be non-nil pointer, if no pointer you get invalidunmarhshalerror at runtime
	//We are decoding into a struct and exporting them. If no matching names, Go attempts to find em

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//dump into input struct in http response
	fmt.Fprintf(w, "%+v\n", input)
}

// GET showmoviehandler endpoint, for now gets id param from current URL and give response
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	// use byname to get value of id param from above slice.
	// if id is invalid, return 404
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r) //goes to errors.go
		return
	}
	// Create new movie struct with ID we got form the URL
	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}
	//Encode struct above to json and punch it
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.servererrorreponse(w, r, err) //Goes to error.Go we set up

	}
}

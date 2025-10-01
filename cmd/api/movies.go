package main

import (
	"errors"
	"fmt"
	"net/http"

	"greenlight.alexedwards.net/internal/data"
	"greenlight.alexedwards.net/internal/validator"
)

// createmovie handler for POST /v1/movies endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	//declare anon struct to hold info we expect to be in the http body
	//Struct will be the *target decode destination
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
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
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()
	//use check() to exec validation checks.
	//adds key and error to err map if check is NOT true

	//use Valid Method to check if any blocks have failed. If they did
	//use failedValidationReponse helper to send reponse to client
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// insert() passes a pointer to validated movie struct.
	// we create a record in the DB and update the struct with new info
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorReponse(w, r, err)
		return
	}

	// send http response with location header to let
	// client know where to find new resource at
	// make empty header map and set new location
	headers := make(http.Header)
	headers.Set("Locaton", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// write JSON response with 201 code status
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorReponse(w, r, err)
	}

	//dump into input struct in http response
	//fmt.Fprintf(w, "%+v\n", input)
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
	//use get method in internal/movies.go to get data for movie
	//also use errors func to check if we return err recordnotfound errr
	//if that happens, return 404 to client.
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorReponse(w, r, err)
		}
		return
	}
	//Encode struct above to json and punch it
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorReponse(w, r, err) //Goes to error.Go we set up

	}
}

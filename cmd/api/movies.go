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

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	//get movieid from URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	//fetch existing movie record from DB, sending a 404 not found if no matching record
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
	//input struct to hold expected data from client
	//added pointers to enable partial updates as before we did not allow nil in fields
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}
	//read Json request body data and put into our above input struct
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//if input.title is nil then we know no title keyword provided by req
	//as .title is now a pointer to a string, we can use the * operator to get the value
	if input.Title != nil {
		movie.Title = *input.Title
	}
	//do same for other fields in the input struct
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}
	//copy stuff form request body to fields of movie record

	//Validate the movie record, send 422 unprocessable if ANY check fails here
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	//pass updated movie record into our new Update() method
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorReponse(w, r, err)
		}
		return
	}
	//final step
	//write updated movie record in JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorReponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	//get movie ID from URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	//delete movie from DB, send 404 if no record found
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorReponse(w, r, err)
		}
		return
	}
	// return 200 if success
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorReponse(w, r, err)
	}
}

// func listmovieHandler for a GET /v1/movies endpoint
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	//define input struct to hold value from request Query string
	var input struct {
		Title        string
		Genres       []string
		data.Filters //new struct found from filters.go
	}
	//ini new validator instance
	v := validator.New()

	//call URL.Query to get map with query string data
	qs := r.URL.Query()

	//use helpers to get title and genres string values
	//use defaults if non provided by client
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	//get page and pagesize values as int
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	//Extract the sort query string value, falling back to ID if not provided
	input.Filters.Sort = app.readString(qs, "sort", "id")
	//supported safelist values for this endpoint
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "runtime"}

	//check validator instance for any errors, act if any
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	//call getall method to get movies, passing in filters if needed
	movies, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorReponse(w, r, err)
		return
	}
	//send JSOn response with all movie data, our main API function
	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies}, nil)
	if err != nil {
		app.serverErrorReponse(w, r, err)
	}

}

package main

//Used for Status codes and some logging

import (
	"fmt"
	"net/http"
)

// Helper for logging error message
func (app *application) logError(r *http.Request, err error) {
	app.logger.Println(err)
}

// The error messages to client with given error code

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	//write response sig writejson helper, if error returned log it
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// servererrorreponse used when app find problems at runtime, logs detailed message
// does send a 500 reponse and JSON
func (app *application) serverErrorReponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "Server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)

}

// not found reponse used to send 404
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// Method no allowed sends our 405 status code
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// special helper func for a 400 bad request message
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// 422 unprocessable entity response
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

package main

import (
	"net/http"
)

//
//declare heandler with plaintxt response about the app status and version

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create map to hold info we want to send in reponse
	//use new envelop map with data inside for response. This way
	//env and version data are nested in response
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"enviroment": app.config.env,
			"version":    version,
		},
	}
	//pass map to jsonmarshalfunc which returns a byte slice wiht encoded JSON
	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorReponse(w, r, err) //now sent to errors.go
	}
}

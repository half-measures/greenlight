package main

import (
	"net/http"
)

//
//declare heandler with plaintxt response about the app status and version

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create map to hold info we want to send in reponse
	data := map[string]string{
		"status":     "available",
		"enviroment": app.config.env,
		"version":    version,
	}
	//pass map to jsonmarshalfunc which returns a byte slice wiht encoded JSON
	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encountered a problem and could not process your request.", http.StatusInternalServerError)
	}

}

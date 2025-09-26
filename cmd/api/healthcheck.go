package main

import (
	"fmt"
	"net/http"
)

//
//declare heandler with plaintxt response about the app status and ver

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "enviroment: %s\n", app.config.env)
	fmt.Fprintf(w, "version: %s\n", version)
}

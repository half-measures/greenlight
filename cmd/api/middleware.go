package main

//This is to avoid panic issues. Currently we panic, close the http
//and log a err message and stack trace
//with below we want to do the same, but send a 500 err
import (
	"fmt"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//create defered function which always run in even of panic
		defer func() {
			//check if theres been a panic, if not, who cares
			if err := recover(); err != nil {
				//if panic found, set connection: close header on response
				w.Header().Set("Connection", "close")
				//value returned by recover is type interface{}
				//so normalize it with fmt.Errorf
				app.serverErrorReponse(w, r, fmt.Errorf("%s", err))
			} //also logs the err for custom logger at ERROR level and send client internal server err response
		}()
		next.ServeHTTP(w, r)
	})

}

//this is all wrapped in router in routes.go

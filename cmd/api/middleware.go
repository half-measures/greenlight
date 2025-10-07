package main

//This is to avoid panic issues. Currently we panic, close the http
//and log a err message and stack trace
//with below we want to do the same, but send a 500 err
import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
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

// /below is a rate limit middleware, were using global empty bucket method
func (app *application) rateLimit(next http.Handler) http.Handler {
	//init new rate limiter to allow avg of 2 reqs per sec,
	//4 for max burst
	limiter := rate.NewLimiter(2, 4)

	//func returing is closure to close over the limiter var
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//call limiter to see if req is permited, if not then return 429
		if !limiter.Allow() {
			app.rateLimitExceededResponse(w, r) //calling in errors.go
			return
		}
		next.ServeHTTP(w, r)
	})
}

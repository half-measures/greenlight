package main

//This is to avoid panic issues. Currently we panic, close the http
//and log a err message and stack trace
//with below we want to do the same, but send a 500 err
import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

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
	//client struct to hold rate limiter and last time seen
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	) //mutex to hold clients IP address and rate limits respectively

	go func() {
		for {
			time.Sleep(time.Minute)
			//lock mutex to prevent limit checks form happening while we clean
			mu.Lock()
			//loop thru and see when last seen, delete those after a bit
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			//importantly, unlock mutex after cleaning
			mu.Unlock()
		}
	}()

	//func returing is closure to close over the limiter var
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//call limiter to see if req is permited, if not then return 429
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorReponse(w, r, err)
			return
		}
		mu.Lock() //lock to prevent convurrent exec in code
		//check if IP alrady in map, if not init a new rate limit and add it
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
		}
		//update last seen time for client
		clients[ip].lastSeen = time.Now()
		//call allow on rate limiter for current IP
		//if req is not allowed, unlock mutex and send 429 resp
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}
		//Very imp, unlock mutex before calling next handler in chain
		//do not use Defer as it means its not unlocked till downstream
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

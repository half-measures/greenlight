package main

//This is to avoid panic issues. Currently we panic, close the http
//and log a err message and stack trace
//with below we want to do the same, but send a 500 err
import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"greenlight.alexedwards.net/internal/data"
	"greenlight.alexedwards.net/internal/validator"
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
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorReponse(w, r, err)
				return
			}
			mu.Lock() //lock to prevent convurrent exec in code
			//check if IP alrady in map, if not init a new rate limit and add it
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)} //These found in main.go
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
		}
		next.ServeHTTP(w, r)
	})
}

// func meant to give valid Auth header then user struct to store in req cntext
// If no Auth given Anon struct engaged instead.
// If Auth is provided but messedup, 401 is return
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		//ADd vary Auth header to response to tell cache that results may vary
		//get auth header from the request
		//Get value of the Auth header form the request, returns "" if no header found
		authorizationHeader := r.Header.Get("Authorization")
		//If no Auth header found, cntextSetUser helper from context.go helps
		//call next hanlder in chain without executing stuff below
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		//Otherwise Auth header should be in bearer token form
		//if not, return 401 unauth resp
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		//extract auth token from header parts finally
		token := headerParts[1]
		//validate token using validator.go
		v := validator.New()

		//if not valid, use invalidtokenrespnse() helper
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		//get details of user with auth token helper used once more
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorReponse(w, r, err)

			}
			return
		}
		r = app.contextSetUser(r, user)
		//call next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// below this should all requireAuth user before being executed itself.
// Should not be checking if user is active unless we know who they are

// Check to see if user is not anon
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the contextGetUser() helper that we made earlier to retrieve the user
		// information from the request context.
		user := app.contextGetUser(r)

		// If the user is anonymous, then call the authenticationRequiredResponse() to
		// inform the client that they should authenticate before trying again.
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		// If the user is not activated, use the inactiveAccountResponse() helper to
		// inform them that they need to activate their account.
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// Middleware to check three checks
// is req from a Auth Non-Anon activated user who has correct permissions?
func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the user from the request context.
		user := app.contextGetUser(r)

		// Get the slice of permissions for the user.
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorReponse(w, r, err)
			return
		}

		// Check if the slice includes the required permission. If it doesn't, then
		// return a 403 Forbidden response.
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		// Otherwise they have the required permission so we call the next handler in
		// the chain.
		next.ServeHTTP(w, r)
	}

	// Wrap this with the requireActivatedUser() middleware before returning it.
	return app.requireActivatedUser(fn)
}

// to prevent cross site scripting ability via Javascript, we want all from the same orgin
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Vary", "Origin")
		//get orgin from header
		origin := r.Header.Get("Origin")
		//only run this part if Orgin req present and trusted orgin is given in CLI on main.go
		if origin != "" && len(app.config.cors.trustedOrigins) != 0 {
			//loop thru trusted orgins if mutliple
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					//if match, set response header with req orgin as value
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

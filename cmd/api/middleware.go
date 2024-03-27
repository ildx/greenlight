package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ildx/greenlight/internal/data"
	"github.com/ildx/greenlight/internal/validator"
	"golang.org/x/time/rate"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add "vary" header to indicate to any caches that the response
		// may vary based on the Authorization header in the request
		w.Header().Add("Vary", "Authorization")

		// retrieve Authorization header; is an empty string if no header found
		authorizationHeader := r.Header.Get("Authorization")

		// if no Authorization header, use contextSetUser() to add
		// an AnonymousUser to context. then call next handler and return
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// expect "Bearer <token>"
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.New()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// retrieve user details with token
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// add user to request context
		r = app.contextSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer function to catch and handle panics
		defer func() {
			// use built-in recover function to catch and handle panics
			if err := recover(); err != nil {
				// if panic, set "Connection: close" header.
				// this will auto-close current connection after response is sent
				w.Header().Set("Connection", "close")
				// normalize the any type to an error
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// client struct to hold the rate limiter and last seen time
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// define mutex and map to store client ips and their rate limits
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// launch go routine to clean up old entries from clients every minute
	go func() {
		for {
			time.Sleep(time.Minute)

			// lock mutex until the cleanup is done
			mu.Lock()

			// loop clients; if no activity in last 3 minutes, delete the entry
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// done, unlock mutex
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// skip rate limiter if not enabled
		if app.config.limiter.enabled {

			// extract client ip from request
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			// lock to prevent concurrent access to the map
			mu.Lock()

			// check if ip is already in map. if not,
			// init new rate limiter and add it with the ip to the map
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}

			// update the last seen time for the client
			clients[ip].lastSeen = time.Now()

			// check if the rate limiter allows the request
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// unlock the mutex.
			// note that unlock is not deferred because we might need to wait
			// until all handlers are done
			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

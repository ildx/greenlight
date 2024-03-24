package main

import (
	"fmt"
	"net/http"
)

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

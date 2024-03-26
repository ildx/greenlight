package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// server settings
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	go func() {
		// quit channel which carries os.Signal values
		quit := make(chan os.Signal, 1)

		// catch SIGINT and SIGTERM signals to channel
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// read from quit channel; this blocks until a signal is received
		s := <-quit

		// log message to indicate signal received.
		// String() is used to to get the signal name
		app.logger.Info("caught signal", "signal", s.String())

		// Exit with status code 0 (success)
		os.Exit(0)
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	return srv.ListenAndServe()
}

package main

import (
	"context"
	"errors"
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

	// shutdown channel to receive errors returned by graceful shutdown
	shutdownError := make(chan error)

	go func() {
		// quit channel which carries os.Signal values
		quit := make(chan os.Signal, 1)

		// catch SIGINT and SIGTERM signals to channel
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// read from quit channel; this blocks until a signal is received
		s := <-quit

		// log message to indicate signal received.
		// String() is used to to get the signal name
		app.logger.Info("shutting down server", "signal", s.String())

		// context with 30s timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// call shutdown and errors to shutdownError channel
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		// log message to indicate that we wait for background routines to complete
		app.logger.Info("completing background tasks", "addr", srv.Addr)

		// block until all routines have completed;
		// return nil to indicate that no error occurred
		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}

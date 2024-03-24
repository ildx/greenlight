package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// global constant holding the version of the app
const version = "1.0.0"

// config struct to hold the configuration of the app
type config struct {
	port int
	env  string
}

// application struct to hold the dependencies
// for the HTTP handlers, helpers, and middleware
type application struct {
	config config
	logger *slog.Logger
}

func main() {
	// declare instance of config
	var cfg config

	// read port and env from the command line
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// init logger for writing to stdout
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// declare app instance
	app := &application{
		config: cfg,
		logger: logger,
	}

	// declare http server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// start server
	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)
	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

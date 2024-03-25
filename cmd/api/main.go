package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// global constant holding the version of the app
const version = "1.0.0"

// config struct to hold the configuration of the app
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
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

	// init logger for writing to stdout
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// read port and env from the command line
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// parse db dsn from env
	err := godotenv.Load(".env")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DB_DSN"), "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	flag.Parse()

	// connect to the database
	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

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

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB returns a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	// create an empty connection pool
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	// create context with a 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ping the database
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

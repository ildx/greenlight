package data

import (
	"database/sql"
	"errors"
)

// error for trying to get a movie that doesn't exist
var (
  ErrRecordNotFound = errors.New("record not found") 
)

// Models wrapper
type Models struct {
  Movies MovieModel
}

// returns a Models struct containing initialized model
func NewModels(db *sql.DB) Models {
  return Models{
    Movies: MovieModel{DB: db},
  }
}

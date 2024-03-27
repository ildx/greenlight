package data

import (
	"database/sql"
	"errors"
)

var (
	// error for trying to get a movie that doesn't exist
	ErrRecordNotFound = errors.New("record not found")

	// error for two users trying to update the same movie at the same time
	ErrEditConflict = errors.New("edit conflict")
)

// Models wrapper
type Models struct {
	Movies MovieModel
	Tokens TokenModel
	Users  UserModel
}

// returns a Models struct containing initialized model
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Tokens: TokenModel{DB: db},
		Users:  UserModel{DB: db},
	}
}

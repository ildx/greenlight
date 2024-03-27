package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/ildx/greenlight/internal/validator"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

type TokenModel struct {
	DB *sql.DB
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

func generateToken(userID int64, tll time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(tll),
		Scope:  scope,
	}

	// init a zero-valued byte slice with a length of 16 bytes
	randomBytes := make([]byte, 16)

	// fill the byte slice with random bytes from os CSPRNG
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// encode the byte slice to a base-32-encoded string.
	// maybe padded with '=' at the end, so omit it
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// generate a SHA-256 hash of the plain text token.
	// it returns an array, which we don't need, so convert it to a slice
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `
    INSERT INTO tokens (hash, user_id, expiry, scope)
    VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
    DELETE FROM tokens
    WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}

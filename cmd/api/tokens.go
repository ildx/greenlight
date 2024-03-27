package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ildx/greenlight/internal/data"
	"github.com/ildx/greenlight/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Println("input", input)

	// find user by email
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		fmt.Println("find user error", err)
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// check if input password matches the actual password
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// no match, call invalid credentials response
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	fmt.Println("all good, next handle tokens...")

	// correct password, generate 24-hour token with scope of "authentication"
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		fmt.Println("BOO token error")
		app.serverErrorResponse(w, r, err)
		return
	}

	fmt.Println("all good?")

	// encode token to JSON and send it to client
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

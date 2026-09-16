package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/dariobaldi/halendar_back/internal/validator"
)

func (app *app) createAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if input.Username != "" {
		data.ValidateUsername(v, input.Username)
	} else {
		data.ValidateEmail(v, input.Email)
	}
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	var user *data.User
	if input.Username != "" {
		user, err = app.models.Users.GetByUsername(input.Username)
	} else {
		user, err = app.models.Users.GetByEmail(input.Email)
	}
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// TODO: Update token time duration. Get back to 24h?
	token, err := app.models.Tokens.New(user.ID, 6*24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	data := map[string]interface{}{
		"id":           user.ID,
		"name":         user.Name,
		"username":     user.Username,
		"token":        token.Plaintext,
		"expiry":       token.Expiry,
		"access_level": user.AccessLevel,
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": data}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *app) changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email       string `json:"email"`
		Username    string `json:"username"`
		Password    string `json:"password"`
		NewPassword string `json:"new_password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if input.Username != "" {
		data.ValidateUsername(v, input.Username)
	} else {
		data.ValidateEmail(v, input.Email)
	}
	data.ValidatePasswordPlaintext(v, input.NewPassword)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	var user *data.User
	if input.Username != "" {
		user, err = app.models.Users.GetByUsername(input.Username)
	} else {
		user, err = app.models.Users.GetByEmail(input.Email)
	}
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	err = user.Password.Set(input.NewPassword)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Update(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeAuthentication, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

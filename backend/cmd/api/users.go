package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/dariobaldi/halendar_back/internal/validator"
	"github.com/google/uuid"
)

func (app *app) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Username == "" {
		input.Username, err = data.GenerateRandomToken()
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	user := &data.User{
		ID:        uuid.New(),
		Name:      input.Name,
		Email:     input.Email,
		Username:  input.Username,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrDuplicateUsername):
			v.AddError("username", "this username is already taken")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
			"access_level":    user.AccessLevel,
		}

		err = app.mailer.Send(user.Email, "user_welcome.html", data)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	app.SendHalendarWS(envelope{"type": "users", "refresh": true})
}

func (app *app) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := app.models.Users.GetListAdmin()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"users": users}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *app) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	requester := app.contextGetUser(r)
	var input struct {
		ID          uuid.UUID `json:"id"`
		Name        *string   `json:"name"`
		Email       *string   `json:"email"`
		Username    *string   `json:"username"`
		Activated   *bool     `json:"activated"`
		AccessLevel *int      `json:"access_level"`
		Version     int       `json:"version"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user, err := app.models.Users.Get(input.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Activated != nil {
		user.Activated = *input.Activated
	}
	if input.AccessLevel != nil && user.AccessLevel < requester.AccessLevel {
		err = app.models.Tokens.DeleteAllForUser("all", user.ID)
		if err != nil {
			app.logger.Error("Couldn't delete all tokens from user after access_level change", "detail", err.Error())
		}
		user.AccessLevel = *input.AccessLevel
	}

	v := validator.New()

	v.Check(user.Version == input.Version, "version", "mistmatch version")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	app.SendHalendarWS(envelope{"type": "users", "refresh": true})
}

func (app *app) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	requester := app.contextGetUser(r)
	var input struct {
		UserID    uuid.UUID `json:"user_id"`
		Activated bool      `json:"activated"`
		Version   int       `json:"version"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.models.Users.Get(input.UserID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if user.Version != input.Version {
		app.serverErrorResponse(w, r, fmt.Errorf("could not update user, mistmatch version"))
		return
	} else if user.AccessLevel > requester.AccessLevel {
		app.notPermittedResponse(w, r)
		return
	}

	user.Activated = input.Activated
	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	app.SendHalendarWS(envelope{"type": "users", "refresh": true})
}

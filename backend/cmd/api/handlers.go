package main

import (
	"net/http"

	"github.com/dariobaldi/halendar/internal/response"
)

func (app *application) status(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Status": "OK",
	}

	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) restricted(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Message": "This is a restricted handler",
	}

	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

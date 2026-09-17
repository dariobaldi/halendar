package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/dariobaldi/halendar_back/internal/push"
)

// pushHealthHandler checks that the configured Firebase service account credentials
// are valid.
func (app *app) pushHealthHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.push.Test(r.Context()); err != nil {
		app.errorResponse(w, r, http.StatusServiceUnavailable, err.Error())
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"status": "available"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// testPushHandler sends a test notification to every device registered for the
// caller, so the Firebase wiring can be confirmed end to end from a device in
// hand. The (optional) JSON body overrides the default title/body.
func (app *app) testPushHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, BodyMaxBytes))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	input := struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}{Title: "Halendar", Body: "Push notifications are working."}

	if trimmed := bytes.TrimSpace(body); len(trimmed) > 0 {
		var override struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		}
		if err := json.Unmarshal(trimmed, &override); err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
		if override.Title != "" {
			input.Title = override.Title
		}
		if override.Body != "" {
			input.Body = override.Body
		}
	}

	devices, err := app.models.Devices.GetForUser(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if len(devices) == 0 {
		app.badRequestResponse(w, r, errors.New("no devices registered for this user — call POST /v1/devices first"))
		return
	}

	sent := 0
	var failures []string
	for _, d := range devices {
		err := app.push.Send(r.Context(), d.PushToken, push.Notification{Title: input.Title, Body: input.Body})
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		sent++
	}

	status := http.StatusOK
	if sent == 0 {
		status = http.StatusServiceUnavailable
	}
	if err := app.writeJSON(w, status, envelope{"sent": sent, "of": len(devices), "errors": failures}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

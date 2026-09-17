package main

import (
	"errors"
	"net/http"

	"github.com/dariobaldi/halendar_back/internal/data"
)

var validDevicePlatforms = map[string]bool{"android": true, "ios": true, "web": true}

// registerDeviceHandler upserts the caller's push token, so a later push.Send
// call knows where to deliver notifications for this user's device. Call this
// again whenever the app receives a refreshed token from the platform SDK.
func (app *app) registerDeviceHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		PushToken string `json:"push_token"`
		Platform  string `json:"platform"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.PushToken == "" {
		app.badRequestResponse(w, r, errors.New(`"push_token" field missing`))
		return
	}
	if !validDevicePlatforms[input.Platform] {
		app.badRequestResponse(w, r, errors.New(`"platform" must be one of: android, ios, web`))
		return
	}

	device := &data.Device{UserID: user.ID, PushToken: input.PushToken, Platform: input.Platform}
	if err := app.models.Devices.Upsert(device); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"device": device}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// listDevicesHandler returns every device registered for the caller.
func (app *app) listDevicesHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	devices, err := app.models.Devices.GetForUser(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"devices": devices}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// unregisterDeviceHandler removes a device, e.g. on logout or when the app detects
// its push permission was revoked.
func (app *app) unregisterDeviceHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		PushToken string `json:"push_token"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.PushToken == "" {
		app.badRequestResponse(w, r, errors.New(`"push_token" field missing`))
		return
	}

	err := app.models.Devices.Delete(user.ID, input.PushToken)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"status": "removed"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

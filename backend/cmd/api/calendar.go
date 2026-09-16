package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/dariobaldi/halendar_back/internal/validator"
	"halendar/calendar"
)

// calendarHealthHandler checks that the configured CalDAV connection works.
func (app *app) calendarHealthHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.calendar.Test(r.Context()); err != nil {
		app.errorResponse(w, r, http.StatusServiceUnavailable, err.Error())
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"status": "available"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// listCalendarsHandler returns the names of the calendars that accept events.
func (app *app) listCalendarsHandler(w http.ResponseWriter, r *http.Request) {
	names, err := app.calendar.Calendars(r.Context())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"calendars": names}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// listEventsHandler returns the events overlapping [start, end], sorted by date.
func (app *app) listEventsHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	v := validator.New()
	start := app.readTime(qs, "start", time.Now(), v)
	end := app.readTime(qs, "end", start.AddDate(0, 0, 7), v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	events, err := app.calendar.Events(r.Context(), start, end)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"events": events}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// checkBusyHandler reports whether [start, end] overlaps a non-cancelled event.
func (app *app) checkBusyHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	v := validator.New()
	start := app.readTime(qs, "start", time.Time{}, v)
	end := app.readTime(qs, "end", time.Time{}, v)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	if start.IsZero() || end.IsZero() {
		app.badRequestResponse(w, r, errors.New(`"start" and "end" query parameters required (RFC3339)`))
		return
	}

	busy, conflicts, err := app.calendar.Busy(r.Context(), start, end)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"busy": busy, "conflicts": conflicts}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// addEventsHandler creates one or more events (a single object or a JSON array), updating
// any whose UID already exists.
func (app *app) addEventsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, BodyMaxBytes))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var events []calendar.Event
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		err = json.Unmarshal(trimmed, &events)
	} else {
		var e calendar.Event
		err = json.Unmarshal(trimmed, &e)
		events = []calendar.Event{e}
	}
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	added := make([]calendar.Event, 0, len(events))
	var failures []string
	for _, e := range events {
		event, err := app.calendar.Add(r.Context(), e)
		if err != nil {
			failures = append(failures, e.Title+": "+err.Error())
			continue
		}
		added = append(added, event)
	}

	status := http.StatusCreated
	if len(failures) > 0 {
		status = http.StatusUnprocessableEntity
	}
	if err := app.writeJSON(w, status, envelope{"events": added, "errors": failures}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteEventHandler removes an event by UID (from the given calendar, or the default one).
func (app *app) deleteEventHandler(w http.ResponseWriter, r *http.Request) {
	uid := app.readStringParam(r, "uid")
	if uid == "" {
		app.badRequestResponse(w, r, errors.New("invalid uid parameter"))
		return
	}
	calendarName := r.URL.Query().Get("calendar")

	if err := app.calendar.Delete(r.Context(), uid, calendarName); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"status": "deleted"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

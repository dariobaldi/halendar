package main

import (
	"net/http"

	"halendar/schedule"
)

// proposeScheduleHandler reads the mail at :uid, picks the first free slot it proposes,
// and returns it with the reply that would be sent. Nothing is booked or sent yet.
func (app *app) proposeScheduleHandler(w http.ResponseWriter, r *http.Request) {
	uid, err := app.readUIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	proposal, err := schedule.Propose(r.Context(), app.mailbox, app.calendar, uid)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"proposal": proposal}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// confirmScheduleHandler books the proposal's event and sends its reply. The caller
// must pass back the exact proposal it got from proposeScheduleHandler.
func (app *app) confirmScheduleHandler(w http.ResponseWriter, r *http.Request) {
	var proposal schedule.Proposal
	if err := app.readJSON(w, r, &proposal); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	booked, err := schedule.Confirm(r.Context(), app.mailbox, app.calendar, &proposal)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusCreated, envelope{"event": booked}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/dariobaldi/halendar_back/internal/mail"
	"github.com/dariobaldi/halendar_back/internal/validator"
)

// mailHealthHandler checks that the configured IMAP and SMTP connections work.
func (app *app) mailHealthHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.mailbox.Test(r.Context()); err != nil {
		app.errorResponse(w, r, http.StatusServiceUnavailable, err.Error())
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"status": "available"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// listFoldersHandler lists the account's folders (INBOX, Sent, Drafts, ...).
func (app *app) listFoldersHandler(w http.ResponseWriter, r *http.Request) {
	folders, err := app.mailbox.Folders(r.Context())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"folders": folders}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// listMessagesHandler returns the most recent messages, newest first. With no filter query
// parameters it returns the n most recent messages (?limit=); with any of unread/since/before/
// from/subject/contains it runs a search instead.
func (app *app) listMessagesHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	v := validator.New()

	query := mail.SearchQuery{
		Unread:   app.readBool(qs, "unread", false),
		Since:    app.readTime(qs, "since", time.Time{}, v),
		Before:   app.readTime(qs, "before", time.Time{}, v),
		From:     app.readString(qs, "from", ""),
		Subject:  app.readString(qs, "subject", ""),
		Contains: app.readString(qs, "contains", ""),
		Max:      app.readInt(qs, "limit", 0, v),
	}
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	isSearch := query.Unread || !query.Since.IsZero() || !query.Before.IsZero() ||
		query.From != "" || query.Subject != "" || query.Contains != ""

	var msgs []mail.Message
	var err error
	if isSearch {
		msgs, err = app.mailbox.Search(r.Context(), query)
	} else {
		limit := query.Max
		if limit <= 0 {
			limit = 20
		}
		msgs, err = app.mailbox.Recent(r.Context(), limit)
	}
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"messages": msgs}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// readMessageHandler returns one message by UID.
func (app *app) readMessageHandler(w http.ResponseWriter, r *http.Request) {
	uid, err := app.readUIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	msg, err := app.mailbox.Read(r.Context(), uid)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"message": msg}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// sendMessageHandler sends a new email.
func (app *app) sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	var input mail.Outgoing
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	id, err := app.mailbox.Send(r.Context(), input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusCreated, envelope{"id": id}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// draftMessageHandler saves the mail in the account's Drafts folder instead of sending it.
func (app *app) draftMessageHandler(w http.ResponseWriter, r *http.Request) {
	var input mail.Outgoing
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	folder, err := app.mailbox.SaveDraft(r.Context(), input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusCreated, envelope{"folder": folder}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// replyMessageHandler replies to a received message, in the same conversation thread.
func (app *app) replyMessageHandler(w http.ResponseWriter, r *http.Request) {
	uid, err := app.readUIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var input struct {
		Text string `json:"text"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	msg, err := app.mailbox.Read(r.Context(), uid)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	id, err := app.mailbox.Reply(r.Context(), *msg, input.Text)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusCreated, envelope{"id": id}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// markMessagesReadHandler marks messages as read or unread.
func (app *app) markMessagesReadHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UIDs []uint32 `json:"uids"`
		Read bool     `json:"read"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if len(input.UIDs) == 0 {
		app.badRequestResponse(w, r, errors.New(`"uids" field missing`))
		return
	}

	if err := app.mailbox.MarkRead(r.Context(), input.Read, input.UIDs...); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"status": "updated"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// moveMessagesHandler moves messages to another folder.
func (app *app) moveMessagesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Folder string   `json:"folder"`
		UIDs   []uint32 `json:"uids"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.Folder == "" || len(input.UIDs) == 0 {
		app.badRequestResponse(w, r, errors.New(`"folder" and "uids" fields required`))
		return
	}

	if err := app.mailbox.Move(r.Context(), input.Folder, input.UIDs...); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"status": "moved"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// readUIDParam reads the ":uid" URL parameter as a mail UID.
func (app *app) readUIDParam(r *http.Request) (uint32, error) {
	s := app.readStringParam(r, "uid")
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, errors.New("invalid uid parameter")
	}
	return uint32(n), nil
}

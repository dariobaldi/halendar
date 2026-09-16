package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/dariobaldi/halendar_back/internal/validator"
	"github.com/dariobaldi/halendar_back/internal/websocket"
	"github.com/julienschmidt/httprouter"
)

func (app *app) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	token := params.ByName("token")
	channel := params.ByName("channel")
	permittedChannels := []string{
		"halendar",
		"delivengo_depots",
		"files",
		"home",
		"listings",
		"marketplaces",
		"notes",
		"orders",
		"packageReturns",
		"packing",
		"picking",
		"picking_items",
		"picking_groups",
		"products",
		"scan",
		"tasks",
	}

	v := validator.New()

	data.ValidateTokenPlaintext(v, token)
	v.Check(validator.PermittedValue(channel, permittedChannels...), "channel", "it doesn't exists")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(token, data.ScopeWsToken, true)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidAuthenticationTokenResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	conn, err := websocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		app.logger.Error("error connecting channel", "detail", err.Error())
		return
	}

	wsHub := app.retriveWebSocket(channel)
	client := &websocket.Client{ID: user.ID, Hub: wsHub, Conn: conn, Send: make(chan []byte, 256)}
	wsHub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (app *app) createWsToken(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	token, err := app.models.Tokens.New(user.ID, 1*time.Minute, data.ScopeWsToken)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"token": token.Plaintext}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

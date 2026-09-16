package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/dariobaldi/halendar_back/internal/validator"
	"github.com/dariobaldi/halendar_back/internal/websocket"
	"github.com/google/uuid"

	"github.com/julienschmidt/httprouter"
)

// Spin a go routine in the background with graceful shutdown
func (app *app) background(fn func()) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		fn()
	}()
}

func checkInternetConnection() bool {
	client := http.Client{
		Timeout: time.Second * 3, // Timeout set to 5 seconds
	}
	_, err := client.Get("http://clients3.google.com/generate_204")
	return err == nil
}

// READING PARAMETERS

func (app *app) readIDParam(r *http.Request) (uuid.UUID, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fmt.Errorf("invalid id parameter")
	}

	return id, nil
}

func (app *app) readIntIDParam(r *http.Request) (int, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid id parameter")
	}

	return id, nil
}

func (app *app) readStringParam(r *http.Request, name string) string {
	params := httprouter.ParamsFromContext(r.Context())

	return params.ByName(name)
}

func (app *app) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	return s
}

func (app *app) readBool(qs url.Values, key string, defaultValue bool) bool {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return defaultValue
	}

	return b
}

func (app *app) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

func (app *app) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *app) readUUID(qs url.Values, key string, defaultValue uuid.UUID) uuid.UUID {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	id, err := uuid.Parse(s)
	if err != nil {
		return defaultValue
	}

	return id
}

func (app *app) readTime(qs url.Values, key string, defaultValue time.Time, v *validator.Validator) time.Time {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		v.AddError("key", "error parsing time: "+err.Error())
		return defaultValue
	}
	return t
}

// PARSING FUNCTIONS

func removeNonNumeric(input string) string {
	re := regexp.MustCompile(`[^0-9]`)
	return re.ReplaceAllString(input, "")
}

func ConvertToCents(amount float64) int {
	converted := math.Round(amount * 100)
	return int(converted)
}

func PriceFromString(price string) (int, error) {
	if price == "" {
		return 0, nil
	}
	price = strings.ReplaceAll(price, ",", ".")
	amount, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid price string")
	}
	return ConvertToCents(amount), nil
}

func ParseTimeString(timeString string) time.Time {
	t, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		return time.Time{}
	}

	return t
}

func errorsToString(errs []error) string {
	if len(errs) == 0 {
		return ""
	}

	errStrings := make([]string, len(errs))
	for i, err := range errs {
		if err != nil {
			errStrings[i] = err.Error()
		}
	}

	return fmt.Sprintf("Errors: %s", strings.Join(errStrings, "\n"))
}

func (app *app) retriveWebSocket(key string) *websocket.Hub {
	hub, ok := app.websockets[key]
	if !ok {
		hub = websocket.NewHub()
		app.background(hub.Run)
		app.websockets[key] = hub
	}

	return hub
}

// SendToWebSocket tries to convert any struct/string/etc to JSON and then sends it
// to the websocket Hub provided as an array of bytes.
func (app *app) SendToWebSocket(hub *websocket.Hub, val any) {
	jsonData, err := json.Marshal(val)
	if err != nil {
		app.logger.Error("error creating json for websocket.", "details", err.Error())
	}
	hub.Broadcast <- websocket.Message{UserID: uuid.Nil, Data: []byte(jsonData)}
	hub.LastUpdated = time.Now()
}

func (app *app) UpdateMessageWS(channels []string) {
	var wsHub *websocket.Hub
	for _, channel := range channels {
		wsHub = app.retriveWebSocket(channel)
		app.SendToWebSocket(wsHub, "Updated")
	}
}

func (app *app) SendHalendarWS(message envelope) {
	wsHub := app.retriveWebSocket("halendar")
	app.SendToWebSocket(wsHub, message)
}

func (app *app) SendToWsUser(UserID uuid.UUID, hub *websocket.Hub, val any) {
	jsonData, err := json.Marshal(val)
	if err != nil {
		app.logger.Error("error creating json for websocket.", "details", err.Error())
	}
	hub.Broadcast <- websocket.Message{UserID: UserID, Data: []byte(jsonData)}
	hub.LastUpdated = time.Now()
}

func trimSpaces(str *string) {
	*str = strings.TrimLeftFunc(*str, unicode.IsSpace)
	*str = strings.TrimRightFunc(*str, unicode.IsSpace)
}

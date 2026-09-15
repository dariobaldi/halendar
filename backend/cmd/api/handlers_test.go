package main

import (
	"net/http"
	"testing"

	"github.com/dariobaldi/halendar/internal/assert"
)

func TestStatus(t *testing.T) {
	t.Run("GET renders the status response", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/status", nil)

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusOK)
		assert.Equal(t, res.BodyFields["Status"], "OK")
	})
}

func TestRestrictedBasicAuth(t *testing.T) {
	t.Run("Unauthenticated users get a 401 response including a WWW-Authenticate header", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/restricted-basic-auth", nil)

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
		assert.Equal(t, res.Header.Get("WWW-Authenticate"), `Basic realm="restricted", charset="UTF-8"`)
	})

	t.Run("Authenticated users get a 200 response", func(t *testing.T) {
		app := newTestApplication(t)

		authUsername := "admin"
		authPassword := "placeholder*77"
		authHashedPassword := "$2a$04$MTmOEATIPE7akymfiaOqyuQemmXp6VAY8pn6yRf3Ya5REVK78umcu"

		app.config.basicAuth.username = authUsername
		app.config.basicAuth.hashedPassword = authHashedPassword

		req := newTestRequest(t, http.MethodGet, "/restricted-basic-auth", nil)
		req.SetBasicAuth(authUsername, authPassword)

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusOK)
		assert.Equal(t, res.BodyFields["Message"], "This is a restricted handler")
	})
}

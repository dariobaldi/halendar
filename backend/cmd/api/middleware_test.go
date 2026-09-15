package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/dariobaldi/halendar/internal/assert"
)

func TestRecoverPanic(t *testing.T) {
	t.Run("Allows normal requests to proceed", func(t *testing.T) {
		app := newTestApplication(t)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, app.recoverPanic(next))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
	})

	t.Run("Recovers from panic and sends a 500 response", func(t *testing.T) {
		var buf bytes.Buffer
		app := newTestApplication(t)
		app.logger = slog.New(slog.NewTextHandler(&buf, nil))

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("something went wrong")
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, app.recoverPanic(next))
		assert.Equal(t, res.StatusCode, http.StatusInternalServerError)
		assert.Equal(t, res.BodyFields["Error"], "The server encountered a problem and could not process your request")
	})
}

func TestLogRequest(t *testing.T) {
	t.Run("Logs the request and response details", func(t *testing.T) {
		var buf bytes.Buffer
		app := newTestApplication(t)
		app.logger = slog.New(slog.NewTextHandler(&buf, nil))

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
			w.Write([]byte(`{"Message": "I'm a test teapot"}`))
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, app.logRequest(next))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
		assert.True(t, strings.Contains(buf.String(), "level=INFO"))
		assert.True(t, strings.Contains(buf.String(), "msg=request"))
		assert.True(t, strings.Contains(buf.String(), "request.method=GET"))
		assert.True(t, strings.Contains(buf.String(), "request.url=/test"))
		assert.True(t, strings.Contains(buf.String(), "response.status=418"))
		assert.True(t, strings.Contains(buf.String(), "response.size=32"))
	})
}

func TestRequireBasicAuthentication(t *testing.T) {
	t.Run("Allows user with valid basic auth credentials to proceed", func(t *testing.T) {
		app := newTestApplication(t)
		authUsername := "admin"
		authPassword := "placeholder*77"
		validHashedPassword := "$2a$04$HLvpR86.wXVT.2KHHkUbFe4/ou3wYGnc9FD7VcKaixofed5enOS.W"

		app.config.basicAuth.username = authUsername
		app.config.basicAuth.hashedPassword = validHashedPassword

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.SetBasicAuth(authUsername, authPassword)

		res := send(t, req, app.requireBasicAuthentication(next))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
	})

	t.Run("Sends a 401 response including WWW-Authenticate header for invalid authentication", func(t *testing.T) {
		validUsername := "admin"
		validPassword := "placeholder*77"
		validHashedPassword := "$2a$04$HLvpR86.wXVT.2KHHkUbFe4/ou3wYGnc9FD7VcKaixofed5enOS.W"

		tests := []struct {
			name         string
			setAuth      bool
			authUsername string
			authPassword string
		}{
			{
				name:    "No basic auth credentials provided",
				setAuth: false,
			},
			{
				name:         "Invalid username provided",
				setAuth:      true,
				authUsername: "wronguser",
				authPassword: validPassword,
			},
			{
				name:         "Invalid password provided",
				setAuth:      true,
				authUsername: validUsername,
				authPassword: "wrongpassword",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				app := newTestApplication(t)

				app.config.basicAuth.username = validUsername
				app.config.basicAuth.hashedPassword = validHashedPassword

				next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				})

				req := newTestRequest(t, http.MethodGet, "/test", nil)
				if tt.setAuth {
					req.SetBasicAuth(tt.authUsername, tt.authPassword)
				}

				res := send(t, req, app.requireBasicAuthentication(next))
				assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
				assert.Equal(t, res.Header.Get("WWW-Authenticate"), `Basic realm="restricted", charset="UTF-8"`)
				assert.Equal(t, res.BodyFields["Error"], "You must be authenticated to access this resource")
			})
		}
	})
}

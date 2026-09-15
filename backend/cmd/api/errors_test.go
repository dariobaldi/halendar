package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/dariobaldi/halendar/internal/assert"
	"github.com/dariobaldi/halendar/internal/validator"
)

func TestReportServerError(t *testing.T) {
	t.Run("Logs error with correct details", func(t *testing.T) {
		var buf bytes.Buffer
		app := newTestApplication(t)
		app.logger = slog.New(slog.NewTextHandler(&buf, nil))

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		app.reportServerError(req, errors.New("this is a test error"))
		assert.True(t, strings.Contains(buf.String(), "level=ERROR"))
		assert.True(t, strings.Contains(buf.String(), `msg="this is a test error"`))
		assert.True(t, strings.Contains(buf.String(), "request.method=GET"))
		assert.True(t, strings.Contains(buf.String(), "request.url=/test"))
	})

}

func TestServerError(t *testing.T) {
	t.Run("Logs error and sends a 500 response without exposing error details", func(t *testing.T) {
		var buf bytes.Buffer
		app := newTestApplication(t)
		app.logger = slog.New(slog.NewTextHandler(&buf, nil))

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.serverError(w, r, errors.New("this is a test error"))
		}))

		assert.Equal(t, res.StatusCode, http.StatusInternalServerError)
		assert.Equal(t, res.Header.Get("Content-Type"), "application/json")
		assert.Equal(t, res.BodyFields["Error"], "The server encountered a problem and could not process your request")

		assert.True(t, strings.Contains(buf.String(), "level=ERROR"))
		assert.True(t, strings.Contains(buf.String(), `msg="this is a test error"`))
		assert.True(t, strings.Contains(buf.String(), "request.method=GET"))
		assert.True(t, strings.Contains(buf.String(), "request.url=/test"))
	})

}

func TestNotFound(t *testing.T) {
	t.Run("Sends a 404 response and error message", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, http.HandlerFunc(app.notFound))
		assert.Equal(t, res.StatusCode, http.StatusNotFound)
		assert.Equal(t, res.BodyFields["Error"], "The requested resource could not be found")
	})
}

func TestMethodNotAllowed(t *testing.T) {
	t.Run("Sends a 405 response and error message", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, http.HandlerFunc(app.methodNotAllowed))
		assert.Equal(t, res.StatusCode, http.StatusMethodNotAllowed)
		assert.Equal(t, res.BodyFields["Error"], "The GET method is not supported for this resource")
	})
}

func TestBadRequest(t *testing.T) {
	t.Run("Sends a 400 response including the error message", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.badRequest(w, r, errors.New("this is a baaaad request"))
		}))
		assert.Equal(t, res.StatusCode, http.StatusBadRequest)
		assert.Equal(t, res.BodyFields["Error"], "This is a baaaad request")
	})
}

func TestFailedValidation(t *testing.T) {
	t.Run("Sends a 422 response including the validation failures", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var v validator.Validator
			v.AddError("This is an validation failure message")
			app.failedValidation(w, r, v)
		}))

		assert.Equal(t, res.StatusCode, http.StatusUnprocessableEntity)
		assert.Equal(t, fmt.Sprint(res.BodyFields["Errors"]), "[This is an validation failure message]")
	})
}

func TestBasicAuthenticationRequired(t *testing.T) {
	t.Run("Sends a 401 response including a WWW-Authenticate header", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, http.HandlerFunc(app.basicAuthenticationRequired))
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
		assert.Equal(t, res.Header.Get("WWW-Authenticate"), `Basic realm="restricted", charset="UTF-8"`)
		assert.Equal(t, res.BodyFields["Error"], "You must be authenticated to access this resource")
	})
}

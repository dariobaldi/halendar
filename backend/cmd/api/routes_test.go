package main

import (
	"net/http"
	"testing"

	"github.com/dariobaldi/halendar/internal/assert"
)

func TestRoutes(t *testing.T) {
	t.Run("Sends a 404 response for non-existent routes", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/nonexistent", nil)

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusNotFound)
		assert.Equal(t, res.BodyFields["Error"], "The requested resource could not be found")
	})

	t.Run("Sends a 405 response for routes with a matching route pattern but no matching HTTP method", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodTrace, "/status", nil)

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusMethodNotAllowed)
		assert.Equal(t, res.BodyFields["Error"], "The TRACE method is not supported for this resource")
	})
}

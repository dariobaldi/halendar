package main

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dariobaldi/halendar/internal/assert"
)

func TestBackgroundTask(t *testing.T) {
	t.Run("Background task runs with no errors", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		app := newTestApplication(t)
		app.logger = logger

		req := httptest.NewRequest("GET", "/test", nil)

		executed := false
		fn := func() error {
			executed = true
			return nil
		}

		app.backgroundTask(req, fn)
		app.wg.Wait()

		assert.True(t, executed)
		assert.True(t, len(buf.String()) == 0)
	})

	t.Run("Error in background task", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		app := newTestApplication(t)
		app.logger = logger

		req := httptest.NewRequest("GET", "/test", nil)

		executed := false
		fn := func() error {
			executed = true
			return errors.New("this is a test error")
		}

		app.backgroundTask(req, fn)
		app.wg.Wait()

		assert.True(t, executed)
		assert.True(t, strings.Contains(buf.String(), "level=ERROR"))
		assert.True(t, strings.Contains(buf.String(), `msg="this is a test error"`))
		assert.True(t, strings.Contains(buf.String(), "request.method=GET"))
		assert.True(t, strings.Contains(buf.String(), "request.url=/test"))
	})

	t.Run("Panic in background task", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		app := newTestApplication(t)
		app.logger = logger

		req := httptest.NewRequest("GET", "/test", nil)

		executed := false
		fn := func() error {
			executed = true
			panic("this is a test error")
		}

		app.backgroundTask(req, fn)
		app.wg.Wait()

		assert.True(t, executed)
		assert.True(t, strings.Contains(buf.String(), "level=ERROR"))
		assert.True(t, strings.Contains(buf.String(), `msg="this is a test error"`))
		assert.True(t, strings.Contains(buf.String(), "request.method=GET"))
		assert.True(t, strings.Contains(buf.String(), "request.url=/test"))
	})
}

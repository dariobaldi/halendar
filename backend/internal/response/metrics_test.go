package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dariobaldi/halendar/internal/assert"
)

func TestMetricsResponseWriter(t *testing.T) {
	t.Run("Track bytes written correctly", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.Write([]byte("test data"))

		assert.Equal(t, mw.StatusCode, http.StatusOK)
		assert.Equal(t, mw.BytesCount, 9)
	})

	t.Run("Track bytes written correctly for multiple writes", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.Write([]byte("test"))
		mw.Write([]byte(" "))
		mw.Write([]byte("data"))

		assert.Equal(t, mw.StatusCode, http.StatusOK)
		assert.Equal(t, mw.BytesCount, 9)
	})

	t.Run("Track status code correctly", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.WriteHeader(http.StatusTeapot)
		mw.Write([]byte("test data"))

		assert.Equal(t, mw.StatusCode, http.StatusTeapot)
		assert.Equal(t, mw.BytesCount, 9)
	})

	t.Run("Ignore status code changes after first write", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.WriteHeader(http.StatusCreated)
		mw.WriteHeader(http.StatusTeapot)

		assert.Equal(t, mw.StatusCode, http.StatusCreated)
	})

	t.Run("Write status code to underlying http.ResponseWriter", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.WriteHeader(http.StatusCreated)
		assert.Equal(t, w.Code, http.StatusCreated)
	})

	t.Run("Write headers to underlying http.ResponseWriter", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.Header().Set("Content-Type", "application/json")
		mw.Header().Set("X-Custom", "test-value")
		assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
		assert.Equal(t, w.Header().Get("X-Custom"), "test-value")
	})

	t.Run("Write body to underlying http.ResponseWriter", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.Write([]byte("test data"))
		assert.Equal(t, w.Body.String(), "test data")
	})
}

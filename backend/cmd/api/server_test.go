package main

import (
	"net"

	"strings"

	"testing"

	"github.com/dariobaldi/halendar/internal/assert"
)

func TestServerConfiguration(t *testing.T) {
	t.Run("Default timeouts are reasonable", func(t *testing.T) {
		assert.True(t, defaultIdleTimeout > 0)
		assert.True(t, defaultReadTimeout > 0)
		assert.True(t, defaultWriteTimeout > defaultReadTimeout)

		if defaultShutdownPeriod <= defaultWriteTimeout {
			t.Errorf("default shutdown period %s must be greater than default write timeout %s", defaultShutdownPeriod, defaultWriteTimeout)
		}
	})
}

func TestServeHTTP(t *testing.T) {

	t.Run("Invalid port configuration causes an error", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.httpPort = -1

		err := app.serveHTTP()
		assert.NotNil(t, err)
	})
}

func TestServeAutoHTTPS(t *testing.T) {

	t.Run("Rejects localhost domain", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.autoHTTPS.domain = "localhost"
		app.config.autoHTTPS.email = "test@example.com"
		app.config.autoHTTPS.staging = true

		err := app.serveAutoHTTPS()
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "localhost"))
	})

	t.Run("Rejects localhost with port", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.autoHTTPS.domain = "localhost:8080"
		app.config.autoHTTPS.email = "test@example.com"
		app.config.autoHTTPS.staging = true

		err := app.serveAutoHTTPS()
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "localhost"))
	})
}

func getFreePort(t *testing.T) int {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

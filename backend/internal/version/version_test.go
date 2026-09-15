package version

import (
	"runtime/debug"
	"testing"

	"github.com/dariobaldi/halendar/internal/assert"
)

func TestGet(t *testing.T) {
	t.Run("Returns version string or unavailable", func(t *testing.T) {
		expectedVersion := "unavailable"
		bi, ok := debug.ReadBuildInfo()
		if ok {
			expectedVersion = bi.Main.Version
		}

		version := Get()
		assert.True(t, version != "")
		assert.Equal(t, version, expectedVersion)
	})
}

func TestGetRevision(t *testing.T) {
	t.Run("Returns revision string or unavailable", func(t *testing.T) {

		revision := GetRevision()
		assert.True(t, revision != "")
		assert.True(t, revision == "unavailable" || len(revision) > 7)
	})
}

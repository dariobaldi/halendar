package env

import (
	"testing"

	"github.com/dariobaldi/halendar/internal/assert"
)

func TestGetString(t *testing.T) {
	t.Run("Return environment variable value when exists", func(t *testing.T) {
		t.Setenv("TEST_STRING", "hello world")

		result := GetString("TEST_STRING", "default_value")
		assert.Equal(t, result, "hello world")
	})

	t.Run("Return default value when environment variable does not exist", func(t *testing.T) {
		result := GetString("NONEXISTENT_STRING", "default_value")
		assert.Equal(t, result, "default_value")
	})

	t.Run("Return empty string when environment variable is empty", func(t *testing.T) {
		t.Setenv("EMPTY_STRING", "")

		result := GetString("EMPTY_STRING", "default_value")
		assert.Equal(t, result, "")
	})

	t.Run("Return environment variable with special characters", func(t *testing.T) {
		t.Setenv("SPECIAL_STRING", "hello@world.com:8080/path?param=value")

		result := GetString("SPECIAL_STRING", "default_value")
		assert.Equal(t, result, "hello@world.com:8080/path?param=value")
	})
}

func TestGetInt(t *testing.T) {
	t.Run("Return parsed integer value when environment variable exists", func(t *testing.T) {
		t.Setenv("TEST_INT", "42")

		result := GetInt("TEST_INT", 0)
		assert.Equal(t, result, 42)
	})

	t.Run("Return default value when environment variable does not exist", func(t *testing.T) {
		result := GetInt("NONEXISTENT_INT", 100)
		assert.Equal(t, result, 100)
	})

	t.Run("Return zero when environment variable is zero", func(t *testing.T) {
		t.Setenv("ZERO_INT", "0")

		result := GetInt("ZERO_INT", 99)
		assert.Equal(t, result, 0)
	})

	t.Run("Panic when environment variable contains invalid integer", func(t *testing.T) {
		t.Setenv("INVALID_INT", "not_an_integer")

		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()

		GetInt("INVALID_INT", 0)
	})
}

func TestGetBool(t *testing.T) {
	t.Run("Return true when environment variable is 'true'", func(t *testing.T) {
		t.Setenv("TEST_BOOL", "true")

		result := GetBool("TEST_BOOL", false)
		assert.Equal(t, result, true)
	})

	t.Run("Return false when environment variable is 'false'", func(t *testing.T) {
		t.Setenv("TEST_BOOL", "false")

		result := GetBool("TEST_BOOL", true)
		assert.Equal(t, result, false)
	})

	t.Run("Return true when environment variable is '1'", func(t *testing.T) {
		t.Setenv("ONE_BOOL", "1")

		result := GetBool("ONE_BOOL", false)
		assert.Equal(t, result, true)
	})

	t.Run("Return false when environment variable is '0'", func(t *testing.T) {
		t.Setenv("ZERO_BOOL", "0")

		result := GetBool("ZERO_BOOL", true)
		assert.Equal(t, result, false)
	})

	t.Run("Handle case insensitive boolean values", func(t *testing.T) {
		t.Setenv("UPPER_BOOL", "TRUE")

		result := GetBool("UPPER_BOOL", false)
		assert.Equal(t, result, true)
	})

	t.Run("Return default value when environment variable does not exist", func(t *testing.T) {
		result := GetBool("NONEXISTENT_BOOL", true)
		assert.Equal(t, result, true)
	})

	t.Run("Panic when environment variable contains invalid boolean", func(t *testing.T) {
		t.Setenv("INVALID_BOOL", "not_a_boolean")

		defer func() {
			r := recover()
			assert.NotNil(t, r)
		}()

		GetBool("INVALID_BOOL", false)
	})
}

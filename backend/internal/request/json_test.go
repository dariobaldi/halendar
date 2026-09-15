package request

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dariobaldi/halendar/internal/assert"
)

type testDecodeJSONTarget struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func TestDecodeJSON(t *testing.T) {
	t.Run("Decode valid JSON successfully", func(t *testing.T) {
		jsonBody := `{"name":"John","age":30,"email":"john@example.com"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))

		var target testDecodeJSONTarget
		err := DecodeJSON(httptest.NewRecorder(), req, &target)
		assert.Nil(t, err)
		assert.Equal(t, target.Name, "John")
		assert.Equal(t, target.Age, 30)
		assert.Equal(t, target.Email, "john@example.com")
	})

	t.Run("Allow unknown fields", func(t *testing.T) {
		jsonBody := `{"name":"John","age":30,"email":"john@example.com","unknown_field":"value"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))

		var target testDecodeJSONTarget
		err := DecodeJSON(httptest.NewRecorder(), req, &target)
		assert.Nil(t, err)
		assert.Equal(t, target.Name, "John")
		assert.Equal(t, target.Age, 30)
		assert.Equal(t, target.Email, "john@example.com")
	})

	t.Run("Return error for empty body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(""))

		var target testDecodeJSONTarget
		err := DecodeJSON(httptest.NewRecorder(), req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "body must not be empty")
	})

	t.Run("Return error for JSON that isn't a struct", func(t *testing.T) {
		jsonBody := `"not-a-struct"`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))

		var target testDecodeJSONTarget
		err := DecodeJSON(httptest.NewRecorder(), req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "body contains incorrect JSON type (at character 14)")
	})

	t.Run("Return error for malformed JSON", func(t *testing.T) {
		jsonBody := `{"name":"John","age":"30",}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))

		var target testDecodeJSONTarget
		err := DecodeJSON(httptest.NewRecorder(), req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "body contains badly-formed JSON (at character 27)")
	})

	t.Run("Return error for unexpected EOF", func(t *testing.T) {
		jsonBody := `{"name"`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))

		var target testDecodeJSONTarget
		err := DecodeJSON(httptest.NewRecorder(), req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "body contains badly-formed JSON")
	})

	t.Run("Return error for incorrect JSON type", func(t *testing.T) {
		jsonBody := `{"name":"John","age":"not-a-number","email":"john@example.com"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))

		var target testDecodeJSONTarget
		err := DecodeJSON(httptest.NewRecorder(), req, &target)

		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), `body contains incorrect JSON type for field "age"`)
	})

	t.Run("Return error for body larger than limit", func(t *testing.T) {

		largeValue := strings.Repeat("a", 1_048_577)
		jsonBody := `{"name":"` + largeValue + `"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))

		var target testDecodeJSONTarget
		err := DecodeJSON(httptest.NewRecorder(), req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "body must not be larger than 1048576 bytes")
	})

	t.Run("Return error for multiple JSON values", func(t *testing.T) {
		jsonBody := `{"name":"John","age":30}{"name":"Jane","age":25}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))

		var target testDecodeJSONTarget
		err := DecodeJSON(httptest.NewRecorder(), req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "body must only contain a single JSON value")
	})
}

func TestDecodeJSONStrict(t *testing.T) {
	t.Run("Return error for unknown fields", func(t *testing.T) {
		jsonBody := `{"name":"John","age":30,"email":"john@example.com","unknown_field":"value"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))

		var target testDecodeJSONTarget
		err := DecodeJSONStrict(httptest.NewRecorder(), req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), `body contains unknown key "unknown_field"`)
	})
}

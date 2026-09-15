package validator

import (
	"testing"

	"github.com/dariobaldi/halendar/internal/assert"
)

func TestHasErrors(t *testing.T) {
	t.Run("Return false when no errors", func(t *testing.T) {
		v := Validator{}

		assert.False(t, v.HasErrors())
	})

	t.Run("Return true when has general errors", func(t *testing.T) {
		v := Validator{Errors: []string{"error1"}}

		assert.True(t, v.HasErrors())
	})

	t.Run("Return true when has field errors", func(t *testing.T) {
		v := Validator{FieldErrors: map[string]string{"field1": "error1"}}

		assert.True(t, v.HasErrors())
	})
}

func TestAddError(t *testing.T) {
	t.Run("Add errors to empty validator", func(t *testing.T) {
		var v Validator
		v.AddError("error 1")
		v.AddError("error 2")

		assert.Equal(t, len(v.Errors), 2)
		assert.Equal(t, v.Errors[0], "error 1")
		assert.Equal(t, v.Errors[1], "error 2")
	})
}

func TestAddFieldError(t *testing.T) {
	t.Run("Add field error to empty validator", func(t *testing.T) {
		var v Validator
		v.AddFieldError("username", "username is required")

		assert.Equal(t, len(v.FieldErrors), 1)
		assert.Equal(t, v.FieldErrors["username"], "username is required")
	})

	t.Run("Do not overwrite existing field error", func(t *testing.T) {
		var v Validator
		v.AddFieldError("username", "first error")
		v.AddFieldError("username", "second error")

		assert.Equal(t, len(v.FieldErrors), 1)
		assert.Equal(t, v.FieldErrors["username"], "first error")
	})
}

func TestCheck(t *testing.T) {
	t.Run("Add error when validation fails", func(t *testing.T) {
		var v Validator
		v.Check(false, "check failed")

		assert.Equal(t, len(v.Errors), 1)
		assert.Equal(t, v.Errors[0], "check failed")
	})

	t.Run("Do not add error when validation passes", func(t *testing.T) {
		var v Validator
		v.Check(true, "should not be added")

		assert.Equal(t, len(v.Errors), 0)
	})
}

func TestCheckField(t *testing.T) {
	t.Run("Add field error when validation fails", func(t *testing.T) {
		var v Validator
		v.CheckField(false, "testField", "testField check failed")

		assert.Equal(t, len(v.FieldErrors), 1)
		assert.Equal(t, v.FieldErrors["testField"], "testField check failed")
	})

	t.Run("Do not add field error when validation passes", func(t *testing.T) {
		var v Validator
		v.CheckField(true, "testField", "should not be added")

		assert.Equal(t, len(v.FieldErrors), 0)
	})
}

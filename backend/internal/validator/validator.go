package validator

import (
	"regexp"
	"slices"
	"strings"
)

var (
	EmailRX    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	UsernameRx = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

func MatchesPattern(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}

func IsMobile(countryCode, phone string) bool {
	patterns := map[string]*regexp.Regexp{
		"FR": regexp.MustCompile(`^((\+33|0)[67])(?:[ _.-]?(\d{2})){4}$`),
		"DE": regexp.MustCompile(`^(\+|00)49(\s?\d{10})$`),
		"BE": regexp.MustCompile(`^(\+|00)32(\s?\d{10})$`),
		"AT": regexp.MustCompile(`^(\+|00)43(\s?\d{9})$`),
		"UK": regexp.MustCompile(`^(\+|00)447([3456789]\d)(\s?\d{7})$`),
		"NL": regexp.MustCompile(`^(\+|00)31(\s?\d{9})$`),
		"PT": regexp.MustCompile(`^(\+|00)351(\s?\d{9})$`),
		"IE": regexp.MustCompile(`^(\+|00)353(\s?\d{8})$`),
		"ES": regexp.MustCompile(`^(\+|00)34(\s?\d{9})$`),
		"IT": regexp.MustCompile(`^(\+|00)39(\s?\d{8,11})$`),
		"CH": regexp.MustCompile(`^(\+|00)41(\s?\d{9})$`),
		"HR": regexp.MustCompile(`^(\+|00)385(\s?\d{8})$`),
		"EE": regexp.MustCompile(`^(\+|00)372(\s?\d{7})$`),
		"LT": regexp.MustCompile(`^(\+|00)370(\s?\d{8})$`),
		"PL": regexp.MustCompile(`^(\+|00)48(\s?\d{9})$`),
		"CZ": regexp.MustCompile(`^(\+|00)420(\s?\d{9})$`),
		"SK": regexp.MustCompile(`^(\+|00)421(\s?\d{9})$`),
		"SI": regexp.MustCompile(`^(\+|00)386(\s?\d{8})$`),
		"LU": regexp.MustCompile(`^(\+|00)352(\s?\d{8})$`),
		"LV": regexp.MustCompile(`^(\+|00)371(\s?\d{8})$`),
		"HU": regexp.MustCompile(`^(\+|00)36(\s?\d{8})$`),
	}

	countryCode = strings.ToUpper(countryCode)

	if pattern, exists := patterns[countryCode]; exists {
		return pattern.MatchString(phone)
	}
	return true
}

func isLatin(r rune) bool {
	// Check if the character is in the basic Latin and Latin Extended ranges
	return (r >= 0x0020 && r <= 0x02AF)
}

func ContainsNonLatinChars(s string) bool {
	for _, r := range s {
		if !isLatin(r) {
			return true
		}
	}
	return false
}

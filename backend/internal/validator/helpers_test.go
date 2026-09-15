package validator

import (
	"regexp"
	"strings"
	"testing"

	"github.com/dariobaldi/halendar/internal/assert"
)

func TestNotBlank(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"Valid non-blank string", "hello", true},
		{"String with spaces", "hello world", true},
		{"Empty string", "", false},
		{"Only spaces", "   ", false},
		{"Only tabs", "\t\t", false},
		{"Mixed whitespace", " \t\n ", false},
		{"Single character", "a", true},
		{"String with leading/trailing spaces", " hello ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NotBlank(tt.value)
			assert.Equal(t, result, tt.expected)
		})
	}
}

func TestMinRunes(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		n        int
		expected bool
	}{
		{"String meets minimum length", "hello", 5, true},
		{"String exceeds minimum length", "hello world", 5, true},
		{"String below minimum length", "hi", 5, false},
		{"Empty string with zero minimum", "", 0, true},
		{"Empty string with positive minimum", "", 1, false},
		{"Unicode characters", "café", 4, true},
		{"Unicode characters below minimum", "café", 5, false},
		{"Emoji characters", "🚀🔐", 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MinRunes(tt.value, tt.n)
			assert.Equal(t, result, tt.expected)
		})
	}
}

func TestMaxRunes(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		n        int
		expected bool
	}{
		{"String within maximum length", "hello", 10, true},
		{"String equals maximum length", "hello", 5, true},
		{"String exceeds maximum length", "hello world", 5, false},
		{"Empty string with zero maximum", "", 0, true},
		{"Empty string with positive maximum", "", 10, true},
		{"Unicode characters within limit", "café", 5, true},
		{"Unicode characters exceed limit", "café", 3, false},
		{"Emoji characters", "🚀🔐", 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaxRunes(tt.value, tt.n)
			assert.Equal(t, result, tt.expected)
		})
	}
}

func TestBetween(t *testing.T) {
	t.Run("Integer values", func(t *testing.T) {
		tests := []struct {
			name     string
			value    int
			min      int
			max      int
			expected bool
		}{
			{"Value within range", 5, 1, 10, true},
			{"Value equals minimum", 1, 1, 10, true},
			{"Value equals maximum", 10, 1, 10, true},
			{"Value below range", 0, 1, 10, false},
			{"Value above range", 11, 1, 10, false},
			{"Single value range", 5, 5, 5, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Between(tt.value, tt.min, tt.max)
				assert.Equal(t, result, tt.expected)
			})
		}
	})

	t.Run("String values", func(t *testing.T) {
		tests := []struct {
			name     string
			value    string
			min      string
			max      string
			expected bool
		}{
			{"String within range", "hello", "a", "z", true},
			{"String equals minimum", "a", "a", "z", true},
			{"String equals maximum", "z", "a", "z", true},
			{"String below range", "A", "a", "z", false},
			{"String above range", "zzz", "a", "z", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Between(tt.value, tt.min, tt.max)
				assert.Equal(t, result, tt.expected)
			})
		}
	})

	t.Run("Float values", func(t *testing.T) {
		tests := []struct {
			name     string
			value    float64
			min      float64
			max      float64
			expected bool
		}{
			{"Float within range", 5.5, 1.0, 10.0, true},
			{"Float below range", 0.5, 1.0, 10.0, false},
			{"Float above range", 10.5, 1.0, 10.0, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := Between(tt.value, tt.min, tt.max)
				assert.Equal(t, result, tt.expected)
			})
		}
	})
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		pattern  string
		expected bool
	}{
		{"Simple pattern match", "hello", "^hello$", true},
		{"Pattern no match", "world", "^hello$", false},
		{"Digit pattern match", "123", "^\\d+$", true},
		{"Digit pattern no match", "abc", "^\\d+$", false},
		{"Email-like pattern", "test@example.com", ".*@.*", true},
		{"Case sensitive match", "Hello", "^hello$", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rx := regexp.MustCompile(tt.pattern)
			result := Matches(tt.value, rx)
			assert.Equal(t, result, tt.expected)
		})
	}
}

func TestIn(t *testing.T) {
	t.Run("String values", func(t *testing.T) {
		tests := []struct {
			name     string
			value    string
			safelist []string
			expected bool
		}{
			{"Value in safelist", "apple", []string{"apple", "banana", "cherry"}, true},
			{"Value not in safelist", "grape", []string{"apple", "banana", "cherry"}, false},
			{"Empty safelist", "apple", []string{}, false},
			{"Single item safelist match", "test", []string{"test"}, true},
			{"Single item safelist no match", "other", []string{"test"}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := In(tt.value, tt.safelist...)
				assert.Equal(t, result, tt.expected)
			})
		}
	})

	t.Run("Integer values", func(t *testing.T) {
		tests := []struct {
			name     string
			value    int
			safelist []int
			expected bool
		}{
			{"Integer in safelist", 2, []int{1, 2, 3}, true},
			{"Integer not in safelist", 4, []int{1, 2, 3}, false},
			{"Empty integer safelist", 1, []int{}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := In(tt.value, tt.safelist...)
				assert.Equal(t, result, tt.expected)
			})
		}
	})
}

func TestAllIn(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		safelist []string
		expected bool
	}{
		{"All values in safelist", []string{"apple", "banana"}, []string{"apple", "banana", "cherry"}, true},
		{"Some values not in safelist", []string{"apple", "grape"}, []string{"apple", "banana", "cherry"}, false},
		{"Empty values slice", []string{}, []string{"apple", "banana"}, true},
		{"Single value in safelist", []string{"apple"}, []string{"apple", "banana"}, true},
		{"Single value not in safelist", []string{"grape"}, []string{"apple", "banana"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AllIn(tt.values, tt.safelist...)
			assert.Equal(t, result, tt.expected)
		})
	}
}

func TestNotIn(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		blocklist []string
		expected  bool
	}{
		{"Value not in blocklist", "apple", []string{"grape", "orange"}, true},
		{"Value in blocklist", "apple", []string{"apple", "banana"}, false},
		{"Empty blocklist", "apple", []string{}, true},
		{"Single item blocklist no match", "apple", []string{"banana"}, true},
		{"Single item blocklist match", "apple", []string{"apple"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NotIn(tt.value, tt.blocklist...)
			assert.Equal(t, result, tt.expected)
		})
	}
}

func TestNoDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected bool
	}{
		{"No duplicates", []string{"apple", "banana", "cherry"}, true},
		{"Has duplicates", []string{"apple", "banana", "apple"}, false},
		{"Empty slice", []string{}, true},
		{"Single item", []string{"apple"}, true},
		{"All same items", []string{"apple", "apple", "apple"}, false},
		{"Two duplicates", []string{"apple", "banana", "apple", "cherry", "banana"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NoDuplicates(tt.values)
			assert.Equal(t, result, tt.expected)
		})
	}
}

func TestIsEmail(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"Valid simple email", "test@example.com", true},
		{"Valid email with subdomain", "user@mail.example.com", true},
		{"Valid email with numbers", "user123@example.org", true},
		{"Valid email with special chars", "user.name+tag@example.com", true},
		{"Invalid email no @", "testexample.com", false},
		{"Invalid email no domain", "test@", false},
		{"Invalid email no local", "@example.com", false},
		{"Invalid email multiple @", "test@@example.com", false},
		{"Empty string", "", false},
		{"Just @", "@", false},
		{"Too long email", strings.Repeat("a", 250) + "@example.com", false},
		{"Valid edge case length", strings.Repeat("a", 240) + "@ex.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmail(tt.value)
			assert.Equal(t, result, tt.expected)
		})
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"Valid HTTP URL", "http://example.com", true},
		{"Valid HTTPS URL", "https://example.com", true},
		{"Valid URL with path", "https://example.com/path", true},
		{"Valid URL with query", "https://example.com?query=value", true},
		{"Valid URL with port", "https://example.com:8080", true},
		{"Invalid URL no scheme", "example.com", false},
		{"Invalid URL no host", "https://", false},
		{"Invalid URL scheme only", "https", false},
		{"Empty string", "", false},
		{"Just scheme", "http://", false},
		{"FTP URL", "ftp://files.example.com", true},
		{"Custom scheme", "custom://example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsURL(tt.value)
			assert.Equal(t, result, tt.expected)
		})
	}
}

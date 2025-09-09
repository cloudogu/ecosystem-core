package config

import (
	"testing"
	"unicode"
)

func TestGeneratePassword(t *testing.T) {
	tests := []struct {
		name          string
		length        int
		expectFailure bool
	}{
		{
			name:          "Minimum length",
			length:        4,
			expectFailure: false,
		},
		{
			name:          "Typical length",
			length:        12,
			expectFailure: false,
		},
		{
			name:          "Long password",
			length:        50,
			expectFailure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := adminPasswordGenerator{}
			password := gen.generatePassword(tt.length)

			if tt.expectFailure {
				if password != "" {
					t.Errorf("Expected empty password, got: %v", password)
				}
				return
			}

			// Validate that the password is of the desired length
			if len(password) != tt.length {
				t.Errorf("Expected password of length %d, got %d", tt.length, len(password))
			}

			// Validate presence of at least one lower-case, upper-case, digit, and special character
			var hasLower, hasUpper, hasDigit, hasSpecial bool
			for _, ch := range password {
				switch {
				case unicode.IsLower(ch):
					hasLower = true
				case unicode.IsUpper(ch):
					hasUpper = true
				case unicode.IsDigit(ch):
					hasDigit = true
				case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
					hasSpecial = true
				}
			}

			if !hasLower || !hasUpper || !hasDigit || !hasSpecial {
				t.Errorf("Password does not meet character type requirements: lower=%v, upper=%v, digit=%v, special=%v", hasLower, hasUpper, hasDigit, hasSpecial)
			}
		})
	}
}

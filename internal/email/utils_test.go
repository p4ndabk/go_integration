package email

import "testing"

func TestIsWelcomeSubject(t *testing.T) {
	testCases := []struct {
		subject  string
		expected bool
	}{
		{"Bem-vindo à nossa plataforma!", true},
		{"Bem-vinda ao sistema", true},
		{"Email de boas-vindas", true},
		{"Welcome to our platform", true},
		{"Welcome to NorthFi", true},
		{"Seja bem-vindo ao nosso sistema", true},
		{"Seja bem-vinda!", true},
		{"BEM-VINDO", true}, // case insensitive
		{"Regular email subject", false},
		{"Password reset email", false},
		{"Verification email", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.subject, func(t *testing.T) {
			result := IsWelcomeSubject(tc.subject)
			if result != tc.expected {
				t.Errorf("IsWelcomeSubject(%q) = %v, expected %v", tc.subject, result, tc.expected)
			}
		})
	}
}
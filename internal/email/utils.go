// Package email provides email utilities and helper functions for email processing.
package email

import "strings"

// IsWelcomeSubject checks if the email subject is a welcome email based on common patterns.
// It performs case-insensitive matching against known welcome email subject patterns
// in both Portuguese and English.
//
// Parameters:
//   - subject: Email subject to analyze
//
// Returns true if the subject matches welcome email patterns, false otherwise.
func IsWelcomeSubject(subject string) bool {
	// Convert to lowercase for case-insensitive comparison
	lowerSubject := strings.ToLower(subject)
	
	// Check for common welcome email patterns
	welcomePatterns := []string{
		"bem-vindo",
		"bem-vinda", 
		"boas-vindas",
		"welcome",
		"welcome to",
		"seja bem-vindo",
		"seja bem-vinda",
	}
	
	for _, pattern := range welcomePatterns {
		if strings.Contains(lowerSubject, pattern) {
			return true
		}
	}
	
	return false
}
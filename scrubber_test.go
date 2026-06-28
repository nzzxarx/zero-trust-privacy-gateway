package main

import (
	"strings"
	"testing"
)

// TestMaskPII checks if the regex core accurately replaces sensitive data
func TestMaskPII(t *testing.T) {
	input := "Contact me at nizar@example.com or call +212600000000. Card: 1234-5678-9012-3456"
	expectedEmail := "[EMAIL PROTECTED]"
	expectedPhone := "[PHONE_REDACTED]"
	expectedCard := "[CREDIT_CARD_REDACTED]"

	output := MaskPII(input)

	if !strings.Contains(output, expectedEmail) {
		t.Errorf("Expected output to contain %s, got %s", expectedEmail, output)
	}
	if !strings.Contains(output, expectedPhone) {
		t.Errorf("Expected output to contain %s, got %s", expectedPhone, output)
	}
	if !strings.Contains(output, expectedCard) {
		t.Errorf("Expected output to contain %s, got %s", expectedCard, output)
	}
}

// TestConcurrencySafe verifies that the processing doesn't mismatch indices under concurrent execution
func TestProcessDocumentsConcurrently(t *testing.T) {
	docs := []string{
		"Doc 1 with email@test.com",
		"Doc 2 secure text",
		"Doc 3 card 4111111111111111",
	}

	results := ProcessDocumentsConcurrently(docs, 2)

	if len(results) != len(docs) {
		t.Fatalf("Expected %d results, got %d", len(docs), len(results))
	}

	if !strings.Contains(results[0], "[EMAIL PROTECTED]") {
		t.Errorf("First doc formatting error: %s", results[0])
	}
	if !strings.Contains(results[2], "[CREDIT_CARD_REDACTED]") {
		t.Errorf("Third doc formatting error: %s", results[2])
	}
}

/*
 * Author: Nizar Akkioui
 * Description: Core logic for concurrent PII redaction.
 */

package main

import (
	"regexp"
	"sync"
)

// Pre-compiled RegEx for high performance (O(1) initialization)
var (
	emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	cardRegex  = regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`)
	phoneRegex = regexp.MustCompile(`\+?\d{1,3}[-.\s]?\(?\d{1,4}\)?[-.\s]?\d{1,4}[-.\s]?\d{1,9}`)
)

// MaskPII applies the regex rules to a single document
func MaskPII(text string) string {
	scrubbed := emailRegex.ReplaceAllString(text, "[EMAIL PROTECTED]")
	scrubbed = cardRegex.ReplaceAllString(scrubbed, "[CREDIT_CARD_REDACTED]")
	scrubbed = phoneRegex.ReplaceAllString(scrubbed, "[PHONE_REDACTED]")
	return scrubbed
}

// ProcessDocumentsConcurrently runs the scrubbing logic across multiple cores safely
func ProcessDocumentsConcurrently(docs []string, maxWorkers int) []string {
	if len(docs) == 0 {
		return docs
	}

	results := make([]string, len(docs))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxWorkers) // Bounded concurrency limit

	for i, doc := range docs {
		wg.Add(1)
		go func(index int, content string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire token
			defer func() { <-semaphore }() // Release token

			results[index] = MaskPII(content)
		}(i, doc)
	}

	wg.Wait()
	return results
}

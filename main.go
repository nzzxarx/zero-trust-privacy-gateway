package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// RequestPayload defines the expected JSON structure
type RequestPayload struct {
	Documents []string `json:"documents"`
}

// ResponsePayload defines the API response
type ResponsePayload struct {
	Status        string   `json:"status"`
	ScrubbedDocs  []string `json:"scrubbed_docs"`
	ProcessingSec float64  `json:"processing_time_sec"`
}

// APIKeyMiddleware ensures only authorized clients can access the API
func APIKeyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		expectedKey := os.Getenv("SCRUBBER_API_KEY")
		clientKey := r.Header.Get("X-API-Key")

		if expectedKey == "" || clientKey != expectedKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Unauthorized"}`))
			return
		}
		next.ServeHTTP(w, r)
	}
}

func scrubHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
		return
	}

	start := time.Now()

	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid JSON format"}`))
		return
	}

	// Step 1: Fast Regex Edge Filter (Go Core Engine from scrubber.go)
	regexScrubbed := ProcessDocumentsConcurrently(payload.Documents, 16)

	// Step 2: Advanced Contextual AI Filtering (Python Microservice via watsonx)
	finalScrubbed := make([]string, len(regexScrubbed))
	for i, doc := range regexScrubbed {
		// Call internal Python AI microservice
		pythonURL := "http://localhost:5000/ai/v1/scrub"
		jsonData, _ := json.Marshal(map[string]string{"document": doc})

		resp, err := http.Post(pythonURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			// Fallback Layer: If Python service is offline, keep the Tier-1 regex output
			finalScrubbed[i] = doc
			continue
		}

		var aiResult map[string]string
		json.NewDecoder(resp.Body).Decode(&aiResult)
		resp.Body.Close()

		finalScrubbed[i] = aiResult["scrubbed_content"]
	}

	response := ResponsePayload{
		Status:        "success",
		ScrubbedDocs:  finalScrubbed,
		ProcessingSec: time.Since(start).Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	mux := http.NewServeMux()

	// Dynamic API Endpoints
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/api/v1/scrub", APIKeyMiddleware(scrubHandler))

	// Serving the static folder architecture for the UI Dashboard
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("🔒 Zero-Trust Scrubber API & Dashboard started on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	<-stop
	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

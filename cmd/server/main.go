package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func formatJSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method. Only POST is allowed.", http.StatusMethodNotAllowed)
		log.Printf("Rejected %s request from %s", r.Method, r.RemoteAddr)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing request body from %s: %v", r.RemoteAddr, err)
		}
	}(r.Body)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		log.Printf("Error reading body from %s: %v", r.RemoteAddr, err)
		return
	}

	if len(bodyBytes) == 0 {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		log.Printf("Received empty body from %s", r.RemoteAddr)
		return
	}

	var jsonData interface{}
	err = json.Unmarshal(bodyBytes, &jsonData)
	if err != nil {
		errorMsg := fmt.Sprintf("Invalid JSON provided: %v", err)
		http.Error(w, errorMsg, http.StatusBadRequest)
		log.Printf("Invalid JSON received from %s: %v", r.RemoteAddr, err)
		return
	}

	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		http.Error(w, "Error formatting JSON", http.StatusInternalServerError)
		log.Printf("Error marshaling JSON from %s (should not happen if unmarshal succeeded): %v", r.RemoteAddr, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(prettyJSON)
	if err != nil {
		log.Printf("Error writing response to %s: %v", r.RemoteAddr, err)
	}
	log.Printf("Successfully processed and formatted JSON from %s", r.RemoteAddr)
}

func main() {
	listenAddr := ":8080"

	mux := http.NewServeMux()
	mux.HandleFunc("/formatjson", formatJSONHandler)

	log.Printf("Starting JSON Formatter/Validator server on %s", listenAddr)

	// Start the HTTP server
	err := http.ListenAndServe(listenAddr, mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

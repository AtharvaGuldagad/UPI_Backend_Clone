package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

// This function handles the incoming HTTP request
func transferHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Log the incoming request
	log.Println("Received transfer request at Go Router...")

	// 2. Extract the query parameters (from, to, amount) sent by the user
	queryParams := r.URL.Query()

	// 3. Build the URL for the Spring Boot Ledger (which is running on 8080)
	springBootURL := "http://localhost:8080/api/transfer"
	
	// Create a new request to send to Spring Boot
	req, err := http.NewRequest(http.MethodPost, springBootURL, nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Attach the query parameters to the new request
	req.URL.RawQuery = queryParams.Encode()

	// 4. Fire the request to Spring Boot using Go's built-in HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to reach Ledger Service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close() // Always close the response body in Go to prevent memory leaks!

	// 5. Read the response from Spring Boot
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	// 6. Send that exact response back to the original client
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
	log.Println("Successfully forwarded response back to client.")
}

func main() {
	// Register the handler function to a specific route
	http.HandleFunc("/api/transfer", transferHandler)

	// Start the Go server on port 8081
	fmt.Println("🚀 Go Router starting on http://localhost:8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
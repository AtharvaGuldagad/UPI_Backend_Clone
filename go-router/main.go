package main

//Ai generated, yet to learn GoLang, so expect some weird code. But it works! :)

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// 1. GLOBAL VARIABLES
// context.Background() is required by the Redis library to manage timeouts and cancellations.
var ctx = context.Background()
var rdb *redis.Client

// 2. THE INIT FUNCTION
// Go automatically runs any function named init() exactly ONCE before main() starts.
// It is the perfect place to set up database connections.
func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Pointing to our new Docker container
	})
	log.Println("Connected to Redis Shield...")
}

func transferHandler(w http.ResponseWriter, r *http.Request) {
	// 3. EXTRACT THE IDEMPOTENCY KEY
	// We expect the client to pass this in the HTTP Headers
	idemKey := r.Header.Get("X-Idempotency-Key")
	if idemKey == "" {
		http.Error(w, "Missing X-Idempotency-Key header. Request rejected.", http.StatusBadRequest)
		return
	}

	// 4. CHECK THE SHIELD (REDIS)
	// We ask Redis: "Do you have this key?"
	val, err := rdb.Get(ctx, idemKey).Result()
	if err == nil {
		// err == nil means Redis FOUND the key. This is a duplicate request!
		log.Printf("BLOCKED: Duplicate request detected for key: %s\n", idemKey)
		msg := fmt.Sprintf("Duplicate request. Previous status: %s", val)
		http.Error(w, msg, http.StatusConflict) // Return a 409 Conflict HTTP status
		return
	}

	log.Printf("Key %s is fresh. Forwarding to Ledger...\n", idemKey)

	// --- THE ORIGINAL PROXY LOGIC ---
	queryParams := r.URL.Query()
	springBootURL := "http://localhost:8080/api/transfer"

	req, err := http.NewRequest(http.MethodPost, springBootURL, nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.URL.RawQuery = queryParams.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to reach Ledger Service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	// --------------------------------

	// 5. SECURE THE SHIELD
	// If Java replies with a 200 OK (Transfer successful!), we save the key to Redis.
	// We give it a Time-To-Live (TTL) of 24 hours so our RAM doesn't fill up forever.
	if resp.StatusCode == http.StatusOK {
		err = rdb.Set(ctx, idemKey, "COMPLETED", 24*time.Hour).Err()
		if err != nil {
			log.Println("Warning: Failed to save key to Redis, system is vulnerable to retries.")
		} else {
			log.Printf("Successfully cached key %s in Redis.\n", idemKey)
		}
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func main() {
	http.HandleFunc("/api/transfer", transferHandler)
	fmt.Println("🚀 Go Router & Shield starting on http://localhost:8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"syslog-visualizer/internal/storage"
	"time"
)

func main() {
	fmt.Println("Syslog Visualizer API starting...")

	// Initialize storage backend
	store := storage.NewMemoryStorage()
	defer store.Close()

	// Setup CORS middleware
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/syslogs", handleGetSyslogs(store))

	// Wrap with CORS
	handler := enableCORS(mux)

	port := ":8080"
	log.Printf("API server listening on %s", port)

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal(err)
	}
}

// enableCORS adds CORS headers to responses
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func handleGetSyslogs(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// TODO: Parse query parameters for filtering
		filters := storage.QueryFilters{
			Limit: 100,
		}

		messages, err := store.Query(filters)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}
}

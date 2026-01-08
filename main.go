package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, req)

		log.Printf("%s %s -> %d (%s)",
			req.Method,
			req.URL.Path,
			rec.status,
			time.Since(start),
		)
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// GET /health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	// POST /echo
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var v any
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	})

	handler := logging(mux)

	log.Printf("listening on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

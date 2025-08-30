package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

// apiConfig holds any stateful, in-memory data we need to keep track of.
// Using a struct like this allows us to cleanly pass state to our handlers.
type apiConfig struct {
	// fileserverHits uses the atomic package to safely increment the integer
	// across multiple goroutines (HTTP requests).
	fileserverHits atomic.Int32
}

func main() {
	// Create a new ServeMux.
	mux := http.NewServeMux()

	// Instantiate our API configuration. This holds the state.
	apiCfg := &apiConfig{}

	// Create a file server handler that serves files out of the current directory.
	const filepathRoot = "."
	fileServer := http.FileServer(http.Dir(filepathRoot))
	
	// Wrap the file server handler with our new metrics middleware.
	// All requests to /app/ will now pass through middlewareMetricsInc first.
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	// Register the readiness and metrics handlers.
	mux.HandleFunc("/healthz", handlerReadiness)
	mux.HandleFunc("/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/reset", apiCfg.handlerReset)

	// Create a new http.Server struct.
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server.
	log.Println("Starting server on :8080")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

// middlewareMetricsInc is a middleware that increments the fileserverHits counter.
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// http.HandlerFunc is an adapter that allows us to use an ordinary function
	// as an HTTP handler.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment the counter. .Add(1) is a thread-safe way to do this.
		cfg.fileserverHits.Add(1)
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// handlerMetrics writes the number of requests that have been counted.
// It's a method on *apiConfig so it can access the fileserverHits data.
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	// .Load() is the thread-safe way to read the value.
	hits := cfg.fileserverHits.Load()
	body := fmt.Sprintf("Hits: %d", hits)
	w.Write([]byte(body))
}

// handlerReset resets the fileserverHits counter to 0.
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	// .Store(0) is the thread-safe way to set the value.
	cfg.fileserverHits.Store(0)
}

// handlerReadiness is the handler for the /healthz endpoint.
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}


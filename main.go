package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

// apiConfig holds any stateful, in-memory data we need to keep track of.
type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	// Create a new ServeMux.
	mux := http.NewServeMux()

	// Instantiate our API configuration.
	apiCfg := &apiConfig{}

	// Create a file server handler.
	const filepathRoot = "."
	fileServer := http.FileServer(http.Dir(filepathRoot))

	// Register handlers for static assets and the main app page.
	// This remains under the /app/ path.
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	// Register API endpoints under the /api namespace.
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /api/reset", apiCfg.handlerReset)

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

// middlewareMetricsInc increments the fileserverHits counter.
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// handlerMetrics writes the number of requests that have been counted.
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hits := cfg.fileserverHits.Load()
	body := fmt.Sprintf("Hits: %d", hits)
	w.Write([]byte(body))
}

// handlerReset resets the fileserverHits counter to 0.
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

// handlerReadiness is the handler for the /healthz endpoint.
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}


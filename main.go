package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	// Register handlers for the user-facing app.
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	// Register API endpoints.
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	// Register admin endpoints.
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

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

// handlerValidateChirp checks the length of a chirp and cleans any profanity.
func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	// Define the structure for decoding the request body.
	type parameters struct {
		Body string `json:"body"`
	}
	// Define the structure for the success response.
	type response struct {
		CleanedBody string `json:"cleaned_body"`
	}

	// Decode the JSON from the request body.
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	// Check if the chirp is too long.
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	// Clean profanity from the chirp.
	cleanedBody := cleanProfanity(params.Body)

	// Respond with the cleaned chirp.
	respondWithJSON(w, http.StatusOK, response{
		CleanedBody: cleanedBody,
	})
}

// cleanProfanity replaces profane words in a string with '****'.
func cleanProfanity(text string) string {
	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(text, " ")
	for i, word := range words {
		// Check the lowercase version of the word against the profane list.
		lowerWord := strings.ToLower(word)
		if _, ok := profaneWords[lowerWord]; ok {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

// middlewareMetricsInc increments the fileserverHits counter.
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// handlerMetrics serves an HTML page with the number of hits.
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	htmlTemplate := `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`

	hits := cfg.fileserverHits.Load()
	responseBody := fmt.Sprintf(htmlTemplate, hits)
	w.Write([]byte(responseBody))
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

// respondWithError is a helper to send a JSON error response.
func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{Error: msg})
}

// respondWithJSON is a helper to send a JSON response.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}


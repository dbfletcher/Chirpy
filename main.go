package main

import (
	"log"
	"net/http"
)

func main() {
	// Create a new ServeMux.
	mux := http.NewServeMux()

	// Create a file server handler that serves files out of the current directory.
	fileServer := http.FileServer(http.Dir("."))

	// Register the file server handler to the root path "/".
	// The StripPrefix is important to ensure that the file server looks in the correct directory.
	mux.Handle("/", fileServer)

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


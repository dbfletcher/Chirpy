package main

import (
	"log"
	"net/http"
)

func main() {
	// Create a new ServeMux. A ServeMux is an HTTP request router (or multiplexer).
	// It compares incoming requests against a list of predefined URL paths,
	// and calls the associated handler for the path whenever a match is found.
	mux := http.NewServeMux()

	// Create a new http.Server struct. We pass in the port we want to listen on
	// and the ServeMux we just created.
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// The ListenAndServe method starts the server. It will block until the server
	// is stopped, so it should be the last line in your main function.
	log.Println("Starting server on :8080")
	err := server.ListenAndServe()
	if err != nil {
		// We use log.Fatal to print the error and exit the program.
		log.Fatal(err)
	}
}


package main

import (
	"fmt"
	"hydraulicPress/lib/API"
	"net/http"
)

// Main function to start the server
func main() {
	// Set up the route for the login handler
	http.HandleFunc("/login", API.LoginHandler)
	http.HandleFunc("/signup", API.SignupHandler)
	// Start the server
	fmt.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

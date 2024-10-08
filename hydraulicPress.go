package main

import (
	"context"
	"fmt"
	"hydraulicPress/lib/API"
	workerpool "hydraulicPress/lib/WorkerPool"
	"hydraulicPress/lib/db_funcs"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/handlers"
)

// Main function to start and manage the server
func main() {
	var server *http.Server
	var wg sync.WaitGroup

	for {
		// Start the server in a separate goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			server = StartServer()
		}()

		// Wait for interrupt signal to gracefully shutdown the server
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt)

		// Check for scheduled shutdown every minute
		go ShutdownAndParse(sigs)

		// Block until a signal is received
		<-sigs
		fmt.Println("Shutting down the server...")

		// You can use server.Shutdown() if using http.Server
		if err := server.Shutdown(context.Background()); err != nil {
			fmt.Println("Error shutting down server:", err)
		}

		wg.Wait() // Wait for the server goroutine to finish
		fmt.Println("Server has been shut down. Parsing log file...")

		// Call your log parsing function
		db_funcs.ParseLogFile()

		fmt.Println("Server will restart...")
	}
}

// StartServer sets up and starts the HTTP server
func StartServer() *http.Server {
	// Set up the routes for the API handlers
	pool := workerpool.NewWorkerPool(10)
	http.HandleFunc("/login", API.LoginHandler)
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		API.SignupHandler(w, r, pool)
	})
	http.HandleFunc("/recovery_email", func(w http.ResponseWriter, r *http.Request) {
		API.SendEmail(w, r, pool)
	})
	http.HandleFunc("/home_page", func(w http.ResponseWriter, r *http.Request) {
		API.HomePageHandler(w, r, pool)
	})

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // Allow all origins for testing; adjust for production
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)(http.DefaultServeMux)
	// Create a new HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: corsHandler, // Set the CORS handler here
	}

	// Start the server
	fmt.Println("Starting server on :8080...")
	go func() {
		if err := server.ListenAndServe(); err != nil {
			fmt.Println("Error starting server:", err)
		}
	}()

	return server
}

// ShutdownAndParse checks for scheduled log parsing
func ShutdownAndParse(sigs chan os.Signal) {
	for {
		now := time.Now()
		// Get the current time in EST (Eastern Standard Time)
		loc, err := time.LoadLocation("America/New_York")
		if err != nil {
			fmt.Println("Error loading location:", err)
			return
		}

		// Check if today is Saturday and the time is 9 PM
		if now.In(loc).Weekday() == time.Saturday && now.In(loc).Hour() == 20 && now.In(loc).Minute() == 1 {
			fmt.Println("Scheduled shutdown at 8 PM EST. Parsing log file...")
			// Trigger server shutdown using the signal channel
			sigs <- os.Interrupt // Send interrupt signal
			db_funcs.ParseLogFile()
			time.Sleep(1 * time.Minute)
		}

		// Sleep for 1 minute before checking again
		time.Sleep(30 * time.Second)
	}
}

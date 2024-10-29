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
		go SignalShutdown(sigs)

		// Block until a signal is received
		<-sigs
		fmt.Println("Shutting down the server...")

		// You can use server.Shutdown() if using http.Server
		if err := server.Shutdown(context.Background()); err != nil {
			fmt.Println("Error shutting down server:", err)
		}
		wg.Wait() // Wait for the server goroutine to finish
		fmt.Println("Server will restart...")
		db_funcs.ParseLogFile()
	}
}

// StartServer sets up and starts the HTTP server
func StartServer() *http.Server {
	// Create a new ServeMux to avoid re-registering the same routes
	mux := http.NewServeMux()

	pool := workerpool.NewWorkerPool(10)
	mux.HandleFunc("/login", API.LoginHandler)
	mux.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		API.SignupHandler(w, r, pool)
	})
	mux.HandleFunc("/recovery_email", func(w http.ResponseWriter, r *http.Request) {
		API.SendEmail(w, r, pool)
	})
	mux.HandleFunc("/verify_email", API.VerifyEmail)
	mux.HandleFunc("/home_page", func(w http.ResponseWriter, r *http.Request) {
		API.HomePageHandler(w, r, pool)
	})
	mux.HandleFunc("/navigation", func(w http.ResponseWriter, r *http.Request) {
		API.Navigation(w, r, pool)
	})
	mux.HandleFunc("/myregiment", func(w http.ResponseWriter, r *http.Request) {
		API.GetAllRegData(w, r, pool)
	})
	mux.HandleFunc("/change-pfp", func(w http.ResponseWriter, r *http.Request) {
		API.UploadPfp(w, r, pool)
	})
	mux.HandleFunc("/fetch_profile", func(w http.ResponseWriter, r *http.Request) {
		API.FetchProfile(w, r, pool)
	})
	mux.HandleFunc("/save_profile", API.SavePrefs)

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // Allow all origins for testing; adjust for production
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)(mux) // Use custom ServeMux instead of DefaultServeMux

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

// check for downtime, and downloadfile
func SignalShutdown(sigs chan os.Signal) {
	for {
		now := time.Now()
		// Get the current time in EST (Eastern Standard Time)
		loc, err := time.LoadLocation("America/New_York")
		if err != nil {
			fmt.Println("Error loading location:", err)
			return
		}

		// Check if today is Saturday and the time is 9 PM
		if now.In(loc).Weekday() == time.Saturday && now.In(loc).Hour() == 20 && now.In(loc).Minute() == 00 {
			fmt.Println("Scheduled shutdown at 8 PM EST. Parsing log file...")
			// Trigger server shutdown using the signal channel
			db_funcs.SftpFileDownload()
			fmt.Println("Now waiting for next minute tick... (server still serves requests for this 65 sec window)")
			time.Sleep(65 * time.Second) //make sure it's next minute
			sigs <- os.Interrupt         // Send interrupt signal
		}
	}
}

package main

import (
	"bufio"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"saloon/database"
	"saloon/handlers"
)

//go:embed static
var staticEmbed embed.FS

// loadEnv loads environment variables from a .env file.
func loadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Strip quotes if they surround the value
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			val = val[1 : len(val)-1]
		}
		os.Setenv(key, val)
	}
	log.Println("Successfully loaded environment variables from .env")
}

func main() {
	// Load config from .env if present
	loadEnv()

	// Initialize Database connection (MongoDB)
	if err := database.InitDB(); err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	// Routes Setup
	// Go 1.22+ net/http supports routing patterns with HTTP methods and path parameters.
	
	// API routes
	http.HandleFunc("GET /api/services", handlers.GetServices)
	http.HandleFunc("POST /api/bookings", handlers.CreateBooking)
	
	// Reviews API
	http.HandleFunc("GET /api/reviews", handlers.GetApprovedReviews)
	http.HandleFunc("POST /api/reviews", handlers.SubmitReview)

	// Admin Authentication API
	http.HandleFunc("POST /api/admin/login", handlers.AdminLogin)
	http.HandleFunc("POST /api/admin/logout", handlers.AdminLogout)
	http.HandleFunc("GET /api/admin/check", handlers.AdminCheck)

	// Admin Protected API
	http.HandleFunc("GET /api/admin/bookings", handlers.AuthMiddleware(handlers.AdminGetBookings))
	http.HandleFunc("POST /api/admin/bookings/{id}/status", handlers.AuthMiddleware(handlers.AdminUpdateBookingStatus))
	http.HandleFunc("GET /api/admin/reviews", handlers.AuthMiddleware(handlers.AdminGetReviews))
	http.HandleFunc("POST /api/admin/reviews/{id}/approve", handlers.AuthMiddleware(handlers.AdminUpdateReviewStatus))
	http.HandleFunc("DELETE /api/admin/reviews/{id}", handlers.AuthMiddleware(handlers.AdminDeleteReview))

	// Embedded Static files handler
	staticFS, err := fs.Sub(staticEmbed, "static")
	if err != nil {
		log.Fatalf("Fatal: Failed to load embedded static sub-filesystem: %v", err)
	}
	
	fileServer := http.FileServer(http.FS(staticFS))
	http.Handle("/", fileServer)

	// Server config
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s...", port)
	if os.Getenv("PORT") == "" {
		log.Printf("Landing Page: http://localhost:%s", port)
		log.Printf("Admin Panel: http://localhost:%s/admin.html", port)
	}
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Fatal: Server error: %v", err)
	}
}

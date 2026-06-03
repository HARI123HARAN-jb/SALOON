package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"saloon/database"
)

// Service represents a salon service offering.
type Service struct {
	Name        string `json:"name"`
	Cost        string `json:"cost"`
	Description string `json:"description"`
}

// Global services list based on finalcode.txt
var services = []Service{
	{Name: "Signature Haircut", Cost: "₹200 - ₹500", Description: "Tailored cuts mapped entirely to your head shape and personal style archetype."},
	{Name: "Beard Design", Cost: "₹150 - ₹350", Description: "Precision line-ups, hot towel treatments, and custom luxury beard conditioning oils."},
	{Name: "Facial Therapy", Cost: "₹500+", Description: "Premium multi-step skincare treatments designed to clean pores and restore radiant skin."},
	{Name: "Groom Makeover", Cost: "Wedding Package", Description: "Elite-tier wedding cosmetics and prep to ensure you look picture-perfect on your big day."},
}

// Simple in-memory session store for Admin Authentication.
var (
	sessions      = make(map[string]string) // token -> username
	sessionsMutex sync.RWMutex
)

// Helper to write JSON responses
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

// GetServices returns the list of salon services.
func GetServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, services)
}

// CreateBooking handles appointment bookings from clients.
func CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Service string `json:"service"`
		Date    string `json:"date"`
		Time    string `json:"time"`
		Notes   string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Basic validation
	if req.Name == "" || req.Phone == "" || req.Service == "" || req.Date == "" || req.Time == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Required fields are missing"})
		return
	}

	booking := database.Booking{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Service: req.Service,
		Date:    req.Date,
		Time:    req.Time,
		Notes:   req.Notes,
	}

	id, err := database.InsertBooking(booking)
	if err != nil {
		log.Printf("Database error creating booking: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save booking"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":    "Booking created successfully",
		"booking_id": id.Hex(),
	})
}

// GetApprovedReviews returns only the approved salon reviews.
func GetApprovedReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reviews, err := database.GetApprovedReviews()
	if err != nil {
		log.Printf("Database error fetching reviews: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve reviews"})
		return
	}

	writeJSON(w, http.StatusOK, reviews)
}

// SubmitReview handles new review submissions from customers.
func SubmitReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name   string `json:"name"`
		Rating int    `json:"rating"`
		Text   string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validation
	if req.Name == "" || req.Rating < 1 || req.Rating > 5 || req.Text == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid review inputs"})
		return
	}

	review := database.Review{
		Name:   req.Name,
		Rating: req.Rating,
		Text:   req.Text,
	}

	id, err := database.InsertReview(review)
	if err != nil {
		log.Printf("Database error saving review: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save review"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":   "Review submitted successfully. Pending admin approval.",
		"review_id": id.Hex(),
	})
}

// AdminLogin authenticates administration credentials and issues a cookie.
func AdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	adminUser := os.Getenv("ADMIN_USERNAME")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "admin123"
	}

	if req.Username != adminUser || req.Password != adminPass {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
		return
	}

	// Generate session token
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		log.Printf("Error generating session token: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to login"})
		return
	}
	token := hex.EncodeToString(tokenBytes)

	// Save token in memory
	sessionsMutex.Lock()
	sessions[token] = req.Username
	sessionsMutex.Unlock()

	// Set HttpOnly session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "Login successful"})
}

// AdminLogout invalidates the administrator's session.
func AdminLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		sessionsMutex.Lock()
		delete(sessions, cookie.Value)
		sessionsMutex.Unlock()
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// AdminCheck verifies if there is an active valid admin session.
func AdminCheck(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]bool{"authenticated": false})
		return
	}

	sessionsMutex.RLock()
	_, valid := sessions[cookie.Value]
	sessionsMutex.RUnlock()

	if !valid {
		writeJSON(w, http.StatusUnauthorized, map[string]bool{"authenticated": false})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"authenticated": true})
}

// AuthMiddleware protects admin routes from unauthenticated access.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
			return
		}

		sessionsMutex.RLock()
		username, valid := sessions[cookie.Value]
		sessionsMutex.RUnlock()

		if !valid {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized session"})
			return
		}

		log.Printf("Admin action by user: %s", username)
		next(w, r)
	}
}

// AdminGetBookings lists all bookings, with optional filter parameters.
func AdminGetBookings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statusFilter := r.URL.Query().Get("status")
	bookings, err := database.GetBookings(statusFilter)
	if err != nil {
		log.Printf("Database error loading bookings: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to load bookings"})
		return
	}

	writeJSON(w, http.StatusOK, bookings)
}

// AdminUpdateBookingStatus updates status of a booking.
func AdminUpdateBookingStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bookingID := r.PathValue("id")
	if bookingID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing booking ID"})
		return
	}

	var req struct {
		Status string `json:"status"` // confirmed, completed, cancelled
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	validStatuses := map[string]bool{"pending": true, "confirmed": true, "completed": true, "cancelled": true}
	if !validStatuses[req.Status] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid status value"})
		return
	}

	err := database.UpdateBookingStatus(bookingID, req.Status)
	if err != nil {
		log.Printf("Database error updating booking: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update booking status"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Booking status updated successfully"})
}

// AdminGetReviews lists all customer reviews.
func AdminGetReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reviews, err := database.GetAllReviews()
	if err != nil {
		log.Printf("Database error loading reviews: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to load reviews"})
		return
	}

	writeJSON(w, http.StatusOK, reviews)
}

// AdminUpdateReviewStatus approves or rejects a customer review.
func AdminUpdateReviewStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reviewID := r.PathValue("id")
	if reviewID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing review ID"})
		return
	}

	var req struct {
		Status string `json:"status"` // approved, rejected
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Status != "approved" && req.Status != "rejected" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid review status"})
		return
	}

	err := database.UpdateReviewStatus(reviewID, req.Status)
	if err != nil {
		log.Printf("Database error updating review: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update review status"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Review status updated successfully"})
}

// AdminDeleteReview permanently removes a review.
func AdminDeleteReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reviewID := r.PathValue("id")
	if reviewID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing review ID"})
		return
	}

	err := database.DeleteReview(reviewID)
	if err != nil {
		log.Printf("Database error deleting review: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete review"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Review deleted successfully"})
}

// Package api provides the HTTP server and handlers for the application.
package api

import "time"

// User represents a user in the system.
type User struct {
	// ID is the unique identifier for the user.
	ID string `json:"id"`
	// Username is the user's login name.
	Username string `json:"username"`
	// Email is the user's email address.
	Email string `json:"email"`
	// Role is the user's assigned role.
	Role string `json:"role"`
	// CreatedAt is the timestamp when the user was created.
	CreatedAt time.Time `json:"created_at"`
}

// Session represents an authenticated session response.
type Session struct {
	// Token is the session token.
	Token string `json:"token"`
	// ExpiresAt is the time when the session expires.
	ExpiresAt time.Time `json:"expires_at"`
}

// ErrorResponse represents an error response from the API.
type ErrorResponse struct {
	// Error is the error message.
	Error string `json:"error"`
	// Code is the HTTP status code.
	Code int `json:"code"`
}

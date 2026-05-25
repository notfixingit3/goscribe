// Package api provides the HTTP server and handlers for the application.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/house/goscribe/test/app/pkg/auth"
)

// healthHandler responds with the service health status.
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// loginHandler authenticates a user and returns a session token.
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	session, err := s.auth.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	response := Session{
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// createUserHandler returns an http.Handler for creating a new user.
func (s *Server) createUserHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Role     string `json:"role"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		user := User{
			ID:        req.Username,
			Username:  req.Username,
			Email:     req.Email,
			Role:      req.Role,
			CreatedAt: time.Now().UTC(),
		}

		data, _ := json.Marshal(user)
		if err := s.store.Set("user:"+req.Username, string(data)); err != nil {
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	})
}

// listUsersHandler returns an http.Handler for listing all users.
func (s *Server) listUsersHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		keys, err := s.store.List("user:")
		if err != nil {
			http.Error(w, "failed to list users", http.StatusInternalServerError)
			return
		}

		var users []User
		for _, key := range keys {
			data, err := s.store.Get(key)
			if err != nil {
				continue
			}
			var user User
			if err := json.Unmarshal([]byte(data), &user); err != nil {
				continue
			}
			users = append(users, user)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	})
}

// UserFromContext retrieves the authenticated user from the request context.
func UserFromContext(r *http.Request) *auth.User {
	return auth.UserFromContext(r.Context())
}

// Package auth provides authentication and session management for the application.
package auth

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey int

const userContextKey contextKey = iota

// RequireAuth returns an http.Handler that requires a valid authentication token.
// It extracts the token from the Authorization header (Bearer scheme) and validates it.
func (am *AuthManager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		user, err := am.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns middleware that requires the authenticated user to have the specified role.
// It must be used after RequireAuth.
func (am *AuthManager) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(userContextKey).(*User)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if user.Role != role {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserFromContext retrieves the authenticated user from the request context.
// Returns nil if no user is present.
func UserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(userContextKey).(*User)
	return user
}

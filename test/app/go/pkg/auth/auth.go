// Package auth provides authentication and session management for the application.
package auth

import (
	"errors"
	"sync"
	"time"
)

// User represents an authenticated user in the system.
type User struct {
	// ID is the unique identifier for the user.
	ID string
	// Username is the user's login name.
	Username string
	// Role is the user's assigned role (e.g., "admin", "user").
	Role string
}

// Session represents an active user session.
type Session struct {
	// Token is the unique session token.
	Token string
	// User is the authenticated user associated with this session.
	User User
	// ExpiresAt is the time when the session expires.
	ExpiresAt time.Time
}

// AuthManager handles user authentication and session management.
type AuthManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	users    map[string]string // username -> password (simulated)
}

// NewAuthManager creates a new AuthManager with default users.
func NewAuthManager() *AuthManager {
	return &AuthManager{
		sessions: make(map[string]*Session),
		users: map[string]string{
			"admin": "admin123",
			"user":  "user123",
		},
	}
}

// Login authenticates a user with the given username and password.
// It returns a Session on success or an error if credentials are invalid.
func (am *AuthManager) Login(username, password string) (*Session, error) {
	am.mu.RLock()
	expectedPassword, ok := am.users[username]
	am.mu.RUnlock()

	if !ok || expectedPassword != password {
		return nil, errors.New("invalid username or password")
	}

	session := &Session{
		Token: generateToken(),
		User: User{
			ID:       username,
			Username: username,
			Role:     "user",
		},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if username == "admin" {
		session.User.Role = "admin"
	}

	am.mu.Lock()
	am.sessions[session.Token] = session
	am.mu.Unlock()

	return session, nil
}

// ValidateToken checks if the given token is valid and returns the associated user.
// It returns an error if the token is invalid or expired.
func (am *AuthManager) ValidateToken(token string) (*User, error) {
	am.mu.RLock()
	session, ok := am.sessions[token]
	am.mu.RUnlock()

	if !ok {
		return nil, errors.New("invalid token")
	}

	if time.Now().After(session.ExpiresAt) {
		am.mu.Lock()
		delete(am.sessions, token)
		am.mu.Unlock()
		return nil, errors.New("session expired")
	}

	return &session.User, nil
}

func generateToken() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

// Package api provides the HTTP server and handlers for the application.
package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/house/goscribe/test/app/go/pkg/auth"
	"github.com/house/goscribe/test/app/go/pkg/storage"
)

// Server is the HTTP server for the application.
type Server struct {
	port   int
	auth   *auth.AuthManager
	store  storage.Store
	server *http.Server
}

// NewServer creates a new Server with the given port, auth manager, and store.
func NewServer(port int, authManager *auth.AuthManager, store storage.Store) *Server {
	return &Server{
		port:  port,
		auth:  authManager,
		store: store,
	}
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/login", s.loginHandler)
	mux.Handle("/users", s.auth.RequireAuth(s.listUsersHandler()))
	mux.Handle("/users/create", s.auth.RequireAuth(s.auth.RequireRole("admin")(s.createUserHandler())))

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server with the given context.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

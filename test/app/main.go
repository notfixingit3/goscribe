// Package main provides the entry point for the test application.
// It demonstrates a simple HTTP service with authentication and storage backends.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/house/goscribe/test/app/pkg/api"
	"github.com/house/goscribe/test/app/pkg/auth"
	"github.com/house/goscribe/test/app/pkg/storage"
)

func main() {
	var port int
	var dsn string
	flag.IntVar(&port, "port", 8080, "HTTP server port")
	flag.StringVar(&dsn, "dsn", "", "SQLite DSN (empty for in-memory store)")
	flag.Parse()

	authManager := auth.NewAuthManager()

	var store storage.Store
	if dsn != "" {
		sqliteStore, err := storage.NewSQLiteStore(dsn)
		if err != nil {
			log.Fatalf("failed to open sqlite store: %v", err)
		}
		store = sqliteStore
	} else {
		store = storage.NewMemoryStore()
	}

	server := api.NewServer(port, authManager, store)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	fmt.Printf("Server listening on :%d\n", port)

	<-sigCh
	fmt.Println("\nShutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}

	fmt.Println("Server stopped.")
}

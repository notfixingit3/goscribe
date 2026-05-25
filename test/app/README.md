# Test Application

This is a test fixture application for GoScribe AI documentation generation testing.

## Overview

A simple HTTP service demonstrating multi-package Go architecture with authentication, storage backends, and REST API handlers.

## Structure

- `main.go` - Entry point with flag parsing and graceful shutdown
- `pkg/auth/` - Authentication and session management
- `pkg/storage/` - Storage interface with memory and SQLite implementations
- `pkg/api/` - HTTP server and request handlers

## Running

```bash
go run main.go -port 8080
go run main.go -port 8080 -dsn "test.db"
```

## API Endpoints

- `GET /health` - Health check
- `POST /login` - Authenticate and get session token
- `GET /users` - List users (requires auth)
- `POST /users/create` - Create user (requires admin role)

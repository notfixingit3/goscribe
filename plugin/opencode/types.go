// Package opencode provides the OpenCode plugin client for integrating GoScribe
// documentation generation into IDE and editor environments.
package opencode

import (
	"context"
	"time"
)

// EventType represents the kind of plugin event.
type EventType string

const (
	// EventFileSaved is dispatched when a source file is saved.
	EventFileSaved EventType = "file_saved"
	// EventTrigger is dispatched on an explicit documentation generation trigger.
	EventTrigger EventType = "trigger"
	// EventInit is dispatched when the plugin finishes initialization.
	EventInit EventType = "init"
	// EventShutdown is dispatched when the plugin begins shutdown.
	EventShutdown EventType = "shutdown"
)

// Event represents a plugin event from the OpenCode agent system.
type Event struct {
	Type        EventType
	FilePath    string
	ProjectPath string
	Timestamp   time.Time
	Metadata    map[string]string
}

// PluginState represents the current lifecycle state of the plugin.
type PluginState int

const (
	// StateUninitialized indicates the plugin has not been initialized.
	StateUninitialized PluginState = iota
	// StateInitializing indicates the plugin is currently initializing.
	StateInitializing
	// StateRunning indicates the plugin is active and processing events.
	StateRunning
	// StateShuttingDown indicates the plugin is gracefully shutting down.
	StateShuttingDown
	// StateStopped indicates the plugin has been stopped.
	StateStopped
)

// String returns a human-readable name for the plugin state.
func (s PluginState) String() string {
	switch s {
	case StateUninitialized:
		return "uninitialized"
	case StateInitializing:
		return "initializing"
	case StateRunning:
		return "running"
	case StateShuttingDown:
		return "shutting_down"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// Handler is the callback signature for processing plugin events.
type Handler func(ctx context.Context, event Event) error

// AgentConnection represents a connection to the OpenCode agent system.
type AgentConnection struct {
	ID          string
	Address     string
	ConnectedAt time.Time
}

// GenerationRequest represents a request to generate documentation.
type GenerationRequest struct {
	SourcePath string
	OutputDir  string
	Force      bool
}

// GenerationResponse represents the result of a documentation generation.
type GenerationResponse struct {
	Success        bool
	OutputDir      string
	FilesGenerated int
	Error          string
}

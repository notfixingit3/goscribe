// Package opencode provides the OpenCode plugin client for integrating GoScribe
// documentation generation into IDE and editor environments.
package opencode

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Client is the OpenCode plugin client that manages the connection to the agent system,
// event handling, and documentation generation lifecycle.
type Client struct {
	config     PluginConfig
	state      PluginState
	handlers   map[EventType][]Handler
	connection *AgentConnection
	watcher    *FileWatcher
	mu         sync.RWMutex
	cancelFunc context.CancelFunc
	ctx        context.Context
}

// NewClient creates a new OpenCode plugin client with the given configuration.
func NewClient(cfg PluginConfig) *Client {
	return &Client{
		config:   cfg,
		state:    StateUninitialized,
		handlers: make(map[EventType][]Handler),
	}
}

// Init initializes the plugin client, establishes an agent connection, and
// transitions to the running state. Returns an error if already initialized or disabled.
func (c *Client) Init(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != StateUninitialized {
		return fmt.Errorf("plugin already initialized (state: %s)", c.state.String())
	}

	c.state = StateInitializing

	if !c.config.Enabled {
		c.state = StateStopped
		return fmt.Errorf("plugin is disabled in configuration")
	}

	c.ctx, c.cancelFunc = context.WithCancel(ctx)

	c.connection = &AgentConnection{
		ID:          generateConnectionID(),
		Address:     c.config.AgentAddress,
		ConnectedAt: time.Now(),
	}

	c.state = StateRunning

	// Start file watcher if auto-generation is enabled.
	if c.config.AutoGenerate {
		c.watcher = NewFileWatcher(c)
		if err := c.watcher.Watch(c.ctx, c.config.ProjectPath); err != nil {
			// Log but don't fail initialization; watcher is optional.
			_ = err
		}
	}

	event := Event{
		Type:        EventInit,
		ProjectPath: "",
		Timestamp:   time.Now(),
	}

	c.mu.Unlock()
	c.dispatchEvent(c.ctx, event)
	c.mu.Lock()

	return nil
}

// Run blocks until the plugin context is canceled. Returns the context error on shutdown.
func (c *Client) Run(ctx context.Context) error {
	c.mu.RLock()
	state := c.state
	c.mu.RUnlock()

	if state != StateRunning {
		return fmt.Errorf("plugin not running (state: %s)", state.String())
	}

	<-c.ctx.Done()
	return c.ctx.Err()
}

// Shutdown gracefully stops the plugin, dispatching a shutdown event with a 5-second
// timeout and releasing all resources.
func (c *Client) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == StateStopped || c.state == StateUninitialized {
		return nil
	}

	c.state = StateShuttingDown

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	event := Event{
		Type:      EventShutdown,
		Timestamp: time.Now(),
	}

	c.mu.Unlock()
	c.dispatchEvent(shutdownCtx, event)
	c.mu.Lock()

	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	c.stopWatcher()
	c.connection = nil
	c.state = StateStopped

	return nil
}

func (c *Client) stopWatcher() {
	if c.watcher != nil {
		c.watcher.Stop()
		c.watcher = nil
	}
}

// State returns the current lifecycle state of the plugin.
func (c *Client) State() PluginState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// Config returns a copy of the plugin configuration.
func (c *Client) Config() PluginConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// Connection returns the current agent connection, or nil if not connected.
func (c *Client) Connection() *AgentConnection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connection
}

// RegisterHandler adds a handler function for the given event type.
func (c *Client) RegisterHandler(eventType EventType, handler Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[eventType] = append(c.handlers[eventType], handler)
}

// HandleEvent processes an incoming event by invoking all registered handlers
// for the event type. Returns an error if the plugin is not running or a handler fails.
func (c *Client) HandleEvent(ctx context.Context, event Event) error {
	c.mu.RLock()
	handlers := c.handlers[event.Type]
	state := c.state
	c.mu.RUnlock()

	if state != StateRunning {
		return fmt.Errorf("plugin not running (state: %s)", state.String())
	}

	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			return fmt.Errorf("handler for %s: %w", event.Type, err)
		}
	}

	return nil
}

func (c *Client) dispatchEvent(ctx context.Context, event Event) {
	c.mu.RLock()
	handlers := c.handlers[event.Type]
	c.mu.RUnlock()

	for _, h := range handlers {
		_ = h(ctx, event)
	}
}

func generateConnectionID() string {
	return fmt.Sprintf("conn-%d", time.Now().UnixNano())
}

package opencode

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	cfg := DefaultConfig()
	client := NewClient(cfg)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.State() != StateUninitialized {
		t.Errorf("expected state uninitialized, got %s", client.State().String())
	}
	if client.Config().Enabled != cfg.Enabled {
		t.Error("config mismatch")
	}
}

func TestClient_Init_Success(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false

	client := NewClient(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := client.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if client.State() != StateRunning {
		t.Errorf("expected state running, got %s", client.State().String())
	}

	conn := client.Connection()
	if conn == nil {
		t.Fatal("expected connection after Init")
	}
	if conn.Address != cfg.AgentAddress {
		t.Errorf("expected address %s, got %s", cfg.AgentAddress, conn.Address)
	}
	if conn.ID == "" {
		t.Error("expected non-empty connection ID")
	}
}

func TestClient_Init_Disabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = false

	client := NewClient(cfg)
	ctx := context.Background()

	err := client.Init(ctx)
	if err == nil {
		t.Fatal("expected error for disabled plugin")
	}
	if client.State() != StateStopped {
		t.Errorf("expected state stopped, got %s", client.State().String())
	}
}

func TestClient_Init_AlreadyInitialized(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false

	client := NewClient(cfg)
	ctx := context.Background()

	if err := client.Init(ctx); err != nil {
		t.Fatalf("first Init failed: %v", err)
	}

	err := client.Init(ctx)
	if err == nil {
		t.Fatal("expected error when initializing twice")
	}
	if client.State() != StateRunning {
		t.Errorf("expected state running, got %s", client.State().String())
	}
}

func TestClient_Run_NotRunning(t *testing.T) {
	cfg := DefaultConfig()
	client := NewClient(cfg)
	ctx := context.Background()

	err := client.Run(ctx)
	if err == nil {
		t.Fatal("expected error when Run called before Init")
	}
}

func TestClient_Run_BlocksUntilShutdown(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false

	client := NewClient(cfg)
	ctx := context.Background()

	if err := client.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- client.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected context error from Run after shutdown")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not unblock after Shutdown")
	}
}

func TestClient_Shutdown_Graceful(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false

	client := NewClient(cfg)
	ctx := context.Background()

	if err := client.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if client.State() != StateStopped {
		t.Errorf("expected state stopped, got %s", client.State().String())
	}
	if client.Connection() != nil {
		t.Error("expected nil connection after shutdown")
	}
}

func TestClient_Shutdown_Idempotent(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false

	client := NewClient(cfg)
	ctx := context.Background()

	if err := client.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("first Shutdown failed: %v", err)
	}
	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("second Shutdown failed: %v", err)
	}

	if client.State() != StateStopped {
		t.Errorf("expected state stopped, got %s", client.State().String())
	}
}

func TestClient_Shutdown_Uninitialized(t *testing.T) {
	cfg := DefaultConfig()
	client := NewClient(cfg)
	ctx := context.Background()

	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown on uninitialized client failed: %v", err)
	}
	if client.State() != StateUninitialized {
		t.Errorf("expected state uninitialized, got %s", client.State().String())
	}
}

func TestClient_RegisterHandler(t *testing.T) {
	cfg := DefaultConfig()
	client := NewClient(cfg)

	var called bool
	client.RegisterHandler(EventInit, func(_ context.Context, _ Event) error {
		called = true
		return nil
	})

	cfg2 := DefaultConfig()
	cfg2.Enabled = true
	cfg2.AutoGenerate = false
	client2 := NewClient(cfg2)

	client2.RegisterHandler(EventInit, func(_ context.Context, _ Event) error {
		called = true
		return nil
	})

	ctx := context.Background()
	if err := client2.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Allow async dispatch to run.
	time.Sleep(100 * time.Millisecond)

	if !called {
		t.Error("expected handler to be called during Init")
	}
}

func TestClient_HandleEvent(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false

	client := NewClient(cfg)
	ctx := context.Background()

	if err := client.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	var called bool
	client.RegisterHandler(EventFileSaved, func(_ context.Context, e Event) error {
		called = true
		return nil
	})

	err := client.HandleEvent(ctx, Event{Type: EventFileSaved})
	if err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestClient_HandleEvent_NotRunning(t *testing.T) {
	cfg := DefaultConfig()
	client := NewClient(cfg)
	ctx := context.Background()

	err := client.HandleEvent(ctx, Event{Type: EventFileSaved})
	if err == nil {
		t.Fatal("expected error when HandleEvent called before Init")
	}
}

func TestClient_HandleEvent_HandlerError(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false

	client := NewClient(cfg)
	ctx := context.Background()

	if err := client.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	expectedErr := errors.New("handler error")
	client.RegisterHandler(EventFileSaved, func(_ context.Context, _ Event) error {
		return expectedErr
	})

	err := client.HandleEvent(ctx, Event{Type: EventFileSaved})
	if err == nil {
		t.Fatal("expected error from failing handler")
	}
}

func TestClient_HandleEvent_MultipleHandlers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false

	client := NewClient(cfg)
	ctx := context.Background()

	if err := client.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	var count int32
	client.RegisterHandler(EventFileSaved, func(_ context.Context, _ Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})
	client.RegisterHandler(EventFileSaved, func(_ context.Context, _ Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	if err := client.HandleEvent(ctx, Event{Type: EventFileSaved}); err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("expected 2 handler calls, got %d", count)
	}
}

func TestClient_StateTransitions(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false

	client := NewClient(cfg)

	if client.State() != StateUninitialized {
		t.Errorf("initial state: expected uninitialized, got %s", client.State().String())
	}

	ctx := context.Background()
	if err := client.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if client.State() != StateRunning {
		t.Errorf("after Init: expected running, got %s", client.State().String())
	}

	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
	if client.State() != StateStopped {
		t.Errorf("after Shutdown: expected stopped, got %s", client.State().String())
	}
}

func TestClient_Config(t *testing.T) {
	cfg := DefaultConfig()
	cfg.OutputDir = "custom-docs"
	client := NewClient(cfg)

	got := client.Config()
	if got.OutputDir != "custom-docs" {
		t.Errorf("expected output dir custom-docs, got %s", got.OutputDir)
	}

	got.OutputDir = "mutated"
	if client.Config().OutputDir != "custom-docs" {
		t.Error("Config returned a mutable reference instead of a copy")
	}
}

func TestClient_Init_WithAutoGenerate(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = true
	cfg.ProjectPath = t.TempDir()

	client := NewClient(cfg)
	ctx := context.Background()

	if err := client.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if client.State() != StateRunning {
		t.Errorf("expected state running, got %s", client.State().String())
	}

	// Watcher should have been created.
	if client.watcher == nil {
		t.Error("expected watcher to be created when AutoGenerate is true")
	}

	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
}

func TestClient_DispatchEvent_Concurrent(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false

	client := NewClient(cfg)
	ctx := context.Background()

	if err := client.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	var wg sync.WaitGroup
	var count int32

	client.RegisterHandler(EventFileSaved, func(_ context.Context, _ Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.HandleEvent(ctx, Event{Type: EventFileSaved})
		}()
	}

	wg.Wait()

	if atomic.LoadInt32(&count) != 10 {
		t.Errorf("expected 10 handler calls, got %d", count)
	}
}

func TestClient_Lifecycle_InitRunShutdown(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.AutoGenerate = false

	client := NewClient(cfg)
	ctx := context.Background()

	if err := client.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- client.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected error from Run after shutdown")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after Shutdown")
	}

	if client.State() != StateStopped {
		t.Errorf("expected state stopped, got %s", client.State().String())
	}
}

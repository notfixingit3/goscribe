// GoScribe OpenCode Plugin Example
//
// This program demonstrates how to embed the GoScribe OpenCode plugin
// into a Go application. It loads plugin config, registers event handlers,
// and runs the plugin lifecycle.
//
// Run:
//
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/house/goscribe/plugin/opencode"
)

func main() {
	cfg, err := opencode.LoadConfigFromProject(".")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	client := opencode.NewClient(cfg)

	client.RegisterHandler(opencode.EventInit, func(_ context.Context, _ opencode.Event) error {
		fmt.Println("plugin initialized")
		return nil
	})

	client.RegisterHandler(opencode.EventFileSaved, func(_ context.Context, e opencode.Event) error {
		fmt.Printf("file changed: %s\n", e.FilePath)
		return nil
	})

	client.RegisterHandler(opencode.EventShutdown, func(_ context.Context, _ opencode.Event) error {
		fmt.Println("plugin shutting down")
		return nil
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := client.Init(ctx); err != nil {
		log.Fatalf("init: %v", err)
	}

	fmt.Println("plugin running. press Ctrl+C to stop.")

	if err := client.Run(ctx); err != nil {
		log.Printf("run ended: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}

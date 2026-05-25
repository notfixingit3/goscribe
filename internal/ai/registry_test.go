package ai

import (
	"context"
	"sync"
	"testing"

	"github.com/house/goscribe/pkg/providers"
)

func TestNewProviderRegistry_Empty(t *testing.T) {
	reg := NewProviderRegistry()
	if names := reg.List(); len(names) != 0 {
		t.Fatalf("expected empty registry, got %v", names)
	}
}

func TestProviderRegistry_RegisterAndGet(t *testing.T) {
	reg := NewProviderRegistry()
	factory := func(cfg providers.ProviderConfig) (Provider, error) {
		return newTestMock("response"), nil
	}

	reg.Register("test-provider", factory)

	got, ok := reg.Get("test-provider")
	if !ok {
		t.Fatal("expected to find registered provider")
	}
	if got == nil {
		t.Fatal("expected non-nil factory")
	}

	_, ok = reg.Get("nonexistent")
	if ok {
		t.Fatal("expected not to find unregistered provider")
	}
}

func TestProviderRegistry_DuplicateRegisterPanics(t *testing.T) {
	reg := NewProviderRegistry()
	factory := func(cfg providers.ProviderConfig) (Provider, error) {
		return newTestMock("response"), nil
	}

	reg.Register("dup", factory)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	reg.Register("dup", factory)
}

func TestProviderRegistry_ListSorted(t *testing.T) {
	reg := NewProviderRegistry()
	factory := func(cfg providers.ProviderConfig) (Provider, error) {
		return newTestMock("response"), nil
	}

	reg.Register("zebra", factory)
	reg.Register("alpha", factory)
	reg.Register("middle", factory)

	names := reg.List()
	expected := []string{"alpha", "middle", "zebra"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(names))
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("names[%d] = %q, want %q", i, names[i], name)
		}
	}
}

func TestProviderRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewProviderRegistry()
	factory := func(cfg providers.ProviderConfig) (Provider, error) {
		return newTestMock("response"), nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)

		go func(i int) {
			defer wg.Done()
			reg.Register("provider-"+string(rune(i)), factory)
		}(i)

		go func() {
			defer wg.Done()
			reg.List()
		}()

		go func(i int) {
			defer wg.Done()
			reg.Get("provider-" + string(rune(i)))
		}(i)
	}
	wg.Wait()
}

func TestProviderRegistry_FactoryCreatesProvider(t *testing.T) {
	reg := NewProviderRegistry()
	reg.Register("mock", func(cfg providers.ProviderConfig) (Provider, error) {
		return newTestMock("hello from " + cfg.Name), nil
	})

	factory, ok := reg.Get("mock")
	if !ok {
		t.Fatal("expected to find mock provider")
	}

	provider, err := factory(providers.ProviderConfig{Name: "mock"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := provider.Generate(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "hello from mock" {
		t.Fatalf("expected 'hello from mock', got %q", resp)
	}
}

func TestDefaultRegistry_HasOpenAIAndOllama(t *testing.T) {
	names := RegisteredProviders()

	found := make(map[string]bool)
	for _, n := range names {
		found[n] = true
	}

	if !found["openai"] {
		t.Error("expected openai in default registry")
	}
	if !found["ollama"] {
		t.Error("expected ollama in default registry")
	}
}

func TestGetProviderFactory(t *testing.T) {
	factory, ok := GetProviderFactory("openai")
	if !ok {
		t.Fatal("expected to find openai factory")
	}
	if factory == nil {
		t.Fatal("expected non-nil factory")
	}

	_, ok = GetProviderFactory("nonexistent")
	if ok {
		t.Fatal("expected not to find nonexistent factory")
	}
}

func TestRegisterProvider_AddsToDefaultRegistry(t *testing.T) {
	factory := func(cfg providers.ProviderConfig) (Provider, error) {
		return newTestMock("custom"), nil
	}

	RegisterProvider("custom-test", factory)

	got, ok := GetProviderFactory("custom-test")
	if !ok {
		t.Fatal("expected to find custom-test in default registry")
	}

	provider, err := got(providers.ProviderConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, _ := provider.Generate(context.Background(), "test")
	if resp != "custom" {
		t.Fatalf("expected 'custom', got %q", resp)
	}
}

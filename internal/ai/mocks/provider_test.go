package mocks

import (
	"context"
	"errors"
	"testing"
)

func TestMockProvider_ReturnsResponsesInOrder(t *testing.T) {
	m := NewMockProvider("first", "second", "third")

	got, err := m.Generate(context.Background(), "prompt 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "first" {
		t.Errorf("expected %q, got %q", "first", got)
	}

	got, err = m.Generate(context.Background(), "prompt 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "second" {
		t.Errorf("expected %q, got %q", "second", got)
	}
}

func TestMockProvider_ErrorsWhenResponsesExhausted(t *testing.T) {
	m := NewMockProvider("only one")

	_, _ = m.Generate(context.Background(), "prompt 1")
	_, err := m.Generate(context.Background(), "prompt 2")
	if err == nil {
		t.Fatal("expected error when responses exhausted")
	}
}

func TestMockProvider_RecordsPrompts(t *testing.T) {
	m := NewMockProvider("a", "b")
	_, _ = m.Generate(context.Background(), "hello")
	_, _ = m.Generate(context.Background(), "world")

	prompts := m.Prompts()
	if len(prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(prompts))
	}
	if prompts[0] != "hello" {
		t.Errorf("expected %q, got %q", "hello", prompts[0])
	}
	if prompts[1] != "world" {
		t.Errorf("expected %q, got %q", "world", prompts[1])
	}
}

func TestMockProvider_CallCount(t *testing.T) {
	m := NewMockProvider("a", "b", "c")
	if m.CallCount() != 0 {
		t.Errorf("expected 0 calls, got %d", m.CallCount())
	}
	_, _ = m.Generate(context.Background(), "x")
	_, _ = m.Generate(context.Background(), "y")
	if m.CallCount() != 2 {
		t.Errorf("expected 2 calls, got %d", m.CallCount())
	}
}

func TestMockProvider_SetErrors(t *testing.T) {
	m := NewMockProvider("response")
	m.SetErrors(errors.New("boom"))

	_, err := m.Generate(context.Background(), "prompt")
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected error 'boom', got %v", err)
	}

	got, err := m.Generate(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("unexpected error after error consumed: %v", err)
	}
	if got != "response" {
		t.Errorf("expected %q, got %q", "response", got)
	}
}

func TestMockProvider_Reset(t *testing.T) {
	m := NewMockProvider("a")
	_, _ = m.Generate(context.Background(), "p1")

	m.Reset()

	if m.CallCount() != 0 {
		t.Errorf("expected 0 calls after reset, got %d", m.CallCount())
	}

	got, err := m.Generate(context.Background(), "p2")
	if err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
	if got != "a" {
		t.Errorf("expected %q after reset, got %q", "a", got)
	}
}

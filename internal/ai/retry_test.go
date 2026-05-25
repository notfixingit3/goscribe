package ai

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// testMock is an in-package mock to avoid import cycles with the mocks sub-package.
type testMock struct {
	mu        sync.Mutex
	responses []string
	index     int
	prompts   []string
	errors    []error
	errIndex  int
}

func newTestMock(responses ...string) *testMock {
	return &testMock{responses: responses}
}

func (m *testMock) Generate(_ context.Context, prompt string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.prompts = append(m.prompts, prompt)

	if m.errIndex < len(m.errors) {
		err := m.errors[m.errIndex]
		m.errIndex++
		return "", err
	}

	if m.index >= len(m.responses) {
		return "", fmt.Errorf("no more responses")
	}
	resp := m.responses[m.index]
	m.index++
	return resp, nil
}

func (m *testMock) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.prompts)
}

func (m *testMock) setErrors(errs ...error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = errs
}

func TestRetryProvider_Success(t *testing.T) {
	mock := newTestMock("hello world")
	retry := NewRetryProvider(mock, 3, 10*time.Millisecond)

	result, err := retry.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Fatalf("expected 'hello world', got %q", result)
	}
	if mock.callCount() != 1 {
		t.Fatalf("expected 1 call, got %d", mock.callCount())
	}
}

func TestRetryProvider_RetriesOnTransientError(t *testing.T) {
	mock := newTestMock("success")
	mock.setErrors(
		errors.New("connection refused"),
		errors.New("timeout: i/o timeout"),
	)

	var logs []string
	retry := NewRetryProvider(mock, 3, 1*time.Millisecond, WithVerbose(func(msg string) {
		logs = append(logs, msg)
	}))

	result, err := retry.Generate(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "success" {
		t.Fatalf("expected 'success', got %q", result)
	}
	// 2 errors + 1 success = 3 total calls
	if mock.callCount() != 3 {
		t.Fatalf("expected 3 calls, got %d", mock.callCount())
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 retry logs, got %d", len(logs))
	}
}

func TestRetryProvider_ExhaustsRetries(t *testing.T) {
	mock := newTestMock("won't reach")
	mock.setErrors(
		errors.New("connection refused"),
		errors.New("connection refused"),
		errors.New("connection refused"),
		errors.New("connection refused"),
	)

	retry := NewRetryProvider(mock, 3, 1*time.Millisecond)

	_, err := retry.Generate(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if mock.callCount() != 4 { // 1 initial + 3 retries
		t.Fatalf("expected 4 calls, got %d", mock.callCount())
	}
}

func TestRetryProvider_NoRetryOnAuthError(t *testing.T) {
	mock := newTestMock("won't reach")
	mock.setErrors(errors.New("unauthorized: invalid api key"))

	retry := NewRetryProvider(mock, 3, 1*time.Millisecond)

	_, err := retry.Generate(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should only call once — no retries for auth errors
	if mock.callCount() != 1 {
		t.Fatalf("expected 1 call (no retries for auth), got %d", mock.callCount())
	}
}

func TestRetryProvider_NoRetryOnBadRequest(t *testing.T) {
	mock := newTestMock("won't reach")
	mock.setErrors(errors.New("invalid model: gpt-5"))

	retry := NewRetryProvider(mock, 3, 1*time.Millisecond)

	_, err := retry.Generate(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if mock.callCount() != 1 {
		t.Fatalf("expected 1 call, got %d", mock.callCount())
	}
}

func TestRetryProvider_ZeroRetries(t *testing.T) {
	mock := newTestMock("won't reach")
	mock.setErrors(errors.New("connection refused"))

	retry := NewRetryProvider(mock, 0, 1*time.Millisecond)

	_, err := retry.Generate(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// 0 retries means 1 attempt only
	if mock.callCount() != 1 {
		t.Fatalf("expected 1 call, got %d", mock.callCount())
	}
}

func TestRetryProvider_ContextCancelled(t *testing.T) {
	mock := newTestMock("won't reach")
	mock.setErrors(
		errors.New("connection refused"),
		errors.New("connection refused"),
	)

	retry := NewRetryProvider(mock, 5, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := retry.Generate(ctx, "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		errMsg    string
		retryable bool
	}{
		// Retryable
		{"connection refused", true},
		{"timeout: i/o timeout", true},
		{"context deadline exceeded", true},
		{"rate limit exceeded", true},
		{"429 Too Many Requests", true},
		{"500 Internal Server Error", true},
		{"502 Bad Gateway", true},
		{"503 Service Unavailable", true},
		{"504 Gateway Timeout", true},
		{"server error: internal server error", true},
		{"network is unreachable", true},
		{"unexpected EOF", true},
		{"temporary failure", true},
		{"connection reset by peer", true},

		// Not retryable
		{"unauthorized", false},
		{"forbidden: access denied", false},
		{"invalid api key provided", false},
		{"incorrect api key", false},
		{"invalid x-api-key header", false},
		{"invalid model: gpt-5", false},
		{"model not found: llama99", false},
		{"context length exceeded", false},
		{"some unknown error", false},
	}

	for _, tt := range tests {
		t.Run(tt.errMsg, func(t *testing.T) {
			got := isRetryable(errors.New(tt.errMsg))
			if got != tt.retryable {
				t.Errorf("isRetryable(%q) = %v, want %v", tt.errMsg, got, tt.retryable)
			}
		})
	}
}

func TestRetryProvider_WaitDuration(t *testing.T) {
	retry := NewRetryProvider(newTestMock("x"), 5, 1*time.Second)

	// Attempt 1: 2^0 * 1s = 1s (with jitter)
	d1 := retry.waitDuration(1)
	if d1 < 800*time.Millisecond || d1 > 1200*time.Millisecond {
		t.Errorf("wait duration for attempt 1 outside expected range: %v", d1)
	}

	// Attempt 2: 2^1 * 1s = 2s (with jitter)
	d2 := retry.waitDuration(2)
	if d2 < 1600*time.Millisecond || d2 > 2400*time.Millisecond {
		t.Errorf("wait duration for attempt 2 outside expected range: %v", d2)
	}

	// Attempt 10: should cap at maxWait (30s)
	d10 := retry.waitDuration(10)
	if d10 > 30*time.Second {
		t.Errorf("wait duration for attempt 10 should be capped at 30s: %v", d10)
	}
}

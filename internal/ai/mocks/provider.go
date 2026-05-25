package mocks

import (
	"context"
	"fmt"
	"sync"
)

type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

var _ Provider = (*MockProvider)(nil)

// MockProvider implements ai.Provider for testing.
// It returns canned responses and records all prompts it receives.
type MockProvider struct {
	mu        sync.Mutex
	responses []string
	index     int
	prompts   []string
	errors    []error
	errIndex  int
}

// NewMockProvider creates a MockProvider that returns the given responses in order.
func NewMockProvider(responses ...string) *MockProvider {
	return &MockProvider{
		responses: responses,
	}
}

// Generate returns the next canned response. If all responses are exhausted,
// it returns an error. It records every prompt received.
func (m *MockProvider) Generate(ctx context.Context, prompt string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.prompts = append(m.prompts, prompt)

	if m.errIndex < len(m.errors) {
		err := m.errors[m.errIndex]
		m.errIndex++
		return "", err
	}

	if m.index >= len(m.responses) {
		return "", fmt.Errorf("mock provider: no more responses (called %d times, have %d responses)", m.index+1, len(m.responses))
	}

	resp := m.responses[m.index]
	m.index++
	return resp, nil
}

// SetErrors configures errors to be returned before responses.
// Each error is returned once, in order, before any responses.
func (m *MockProvider) SetErrors(errs ...error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = errs
}

// Prompts returns all prompts the provider has received.
func (m *MockProvider) Prompts() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]string, len(m.prompts))
	copy(cp, m.prompts)
	return cp
}

// CallCount returns the number of times Generate was called.
func (m *MockProvider) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.prompts)
}

// Reset clears all recorded prompts and resets the response index.
func (m *MockProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prompts = nil
	m.index = 0
	m.errors = nil
	m.errIndex = 0
}

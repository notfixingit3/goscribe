package ai

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"
)

// RetryProvider wraps a Provider with exponential backoff retry logic.
// It retries on transient errors such as network timeouts, rate limits,
// and connection failures. Non-retryable errors (auth, bad requests) are
// returned immediately.
type RetryProvider struct {
	inner   Provider
	max     int
	backoff time.Duration
	maxWait time.Duration
	jitter  float64
	verbose func(string)
}

// RetryOption configures a RetryProvider.
type RetryOption func(*RetryProvider)

// WithVerbose sets a callback for retry log messages.
func WithVerbose(fn func(string)) RetryOption {
	return func(r *RetryProvider) { r.verbose = fn }
}

// WithMaxWait sets the maximum backoff duration. Defaults to 30s.
func WithMaxWait(d time.Duration) RetryOption {
	return func(r *RetryProvider) { r.maxWait = d }
}

// NewRetryProvider wraps inner with retry logic.
// max is the number of retry attempts (0 = no retries).
// backoff is the initial delay between retries.
func NewRetryProvider(inner Provider, max int, backoff time.Duration, opts ...RetryOption) *RetryProvider {
	r := &RetryProvider{
		inner:   inner,
		max:     max,
		backoff: backoff,
		maxWait: 30 * time.Second,
		jitter:  0.2,
		verbose: func(string) {},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Generate calls the inner provider's Generate with retry on transient errors.
func (r *RetryProvider) Generate(ctx context.Context, prompt string) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= r.max; attempt++ {
		if attempt > 0 {
			wait := r.waitDuration(attempt)
			r.verbose(fmt.Sprintf("retry %d/%d after %s: %v", attempt, r.max, wait, lastErr))

			select {
			case <-ctx.Done():
				return "", fmt.Errorf("context canceled during retry: %w", ctx.Err())
			case <-time.After(wait):
			}
		}

		result, err := r.inner.Generate(ctx, prompt)
		if err == nil {
			return result, nil
		}

		lastErr = err
		if !isRetryable(err) {
			return "", err
		}
	}

	return "", fmt.Errorf("failed after %d retries: %w", r.max, lastErr)
}

func (r *RetryProvider) waitDuration(attempt int) time.Duration {
	exp := time.Duration(math.Pow(2, float64(attempt-1))) * r.backoff
	if exp > r.maxWait {
		exp = r.maxWait
	}
	// Add jitter to avoid thundering herd
	jitterDelta := time.Duration(float64(exp) * r.jitter)
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(2*jitterDelta+1)))
	exp += time.Duration(n.Int64()) - jitterDelta
	if exp > r.maxWait {
		exp = r.maxWait
	}
	if exp < 0 {
		exp = r.backoff
	}
	return exp
}

// isRetryable determines whether an error is transient and worth retrying.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	// Non-retryable: auth and permission errors
	if strings.Contains(msg, "unauthorized") ||
		strings.Contains(msg, "forbidden") ||
		strings.Contains(msg, "invalid api key") ||
		strings.Contains(msg, "invalid x-api-key") ||
		strings.Contains(msg, "incorrect api key") {
		return false
	}

	// Non-retryable: bad request / validation errors
	if strings.Contains(msg, "invalid model") ||
		strings.Contains(msg, "model not found") ||
		strings.Contains(msg, "context length") {
		return false
	}

	// Retryable: transient failures
	retryablePatterns := []string{
		"timeout",
		"deadline exceeded",
		"connection refused",
		"connection reset",
		"temporary",
		"rate limit",
		"too many requests",
		"429",
		"500",
		"502",
		"503",
		"504",
		"server error",
		"internal server error",
		"bad gateway",
		"service unavailable",
		"gateway timeout",
		"i/o timeout",
		"network",
		"eof",
	}
	for _, pattern := range retryablePatterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}

	// Default: don't retry unknown errors
	return false
}

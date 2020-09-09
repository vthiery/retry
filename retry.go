package retry

import (
	"context"
	"fmt"
	"time"
)

// Retry allows to retry a function until the amount of attempts is exhausted,
// using a backoff mechanism to wait between successive attempts.
// By default, there retrier will:
//     * attempts ad infinitum
//     * not observe wait between successive attempts
//     * retry on all errors
type Retry struct {
	maxAttempts *int
	backoff     Backoff
	policy      Policy
}

// Backoff describes the backoff interface.
type Backoff interface {
	// Next returns the duration to wait before performing the next attempt.
	Next(attempt int) time.Duration
}

// Policy describes the retry policy based on the error.
// It must return true for retryable errors and false otherwise.
type Policy func(err error) bool

// NoAttemptsAllowedError is used to signal that no attempts are allowed.
type NoAttemptsAllowedError struct {
	MaxAttempts int
}

// Error returns the error message.
func (e *NoAttemptsAllowedError) Error() string {
	return fmt.Sprintf("no attempts are allowed with max attempts set to %d", e.MaxAttempts)
}

// Do attempts to execute the function `fn` until the amount of attempts is
// exhausted and wait between attempts according to the backoff strategy
// set on Retry.
func (r *Retry) Do(ctx context.Context, fn func(context.Context) error) error {
	if r.maxAttempts != nil && *r.maxAttempts < 1 {
		return &NoAttemptsAllowedError{
			MaxAttempts: *r.maxAttempts,
		}
	}

	attempt := 0
	for {
		if err := fn(ctx); err != nil {
			if !r.policy(err) {
				return fmt.Errorf("got a non-retryable error: %w", err)
			}

			attempt++
			if r.exhaustedAttempts(attempt) {
				return fmt.Errorf("all attempts have been exhausted, finished with error: %w", err)
			}
			if err := r.waitBackoffTime(ctx, attempt); err != nil {
				return err
			}
			continue
		}
		return nil
	}
}

func (r Retry) exhaustedAttempts(attempt int) bool {
	return r.maxAttempts != nil && attempt >= *r.maxAttempts
}

func (r Retry) waitBackoffTime(ctx context.Context, attempt int) error {
	if r.backoff == nil {
		return ctx.Err()
	}
	return wait(ctx, r.backoff.Next(attempt))
}

func wait(ctx context.Context, duration time.Duration) error {
	waitCtx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-waitCtx.Done():
		return nil
	}
}

// Option is an option to add to the Retry.
type Option func(r *Retry)

// WithMaxAttempts sets a the maximum number of attempts to perform.
func WithMaxAttempts(maxAttempts int) Option {
	return func(r *Retry) {
		r.maxAttempts = &maxAttempts
	}
}

// WithBackoff sets a backoff strategy.
func WithBackoff(backoff Backoff) Option {
	return func(r *Retry) {
		r.backoff = backoff
	}
}

// WithPolicy sets the retry policy.
func WithPolicy(policy Policy) Option {
	return func(r *Retry) {
		r.policy = policy
	}
}

// New creates an instance of Retry.
// By defaults, there is no retry limit and no backoff.
func New(opts ...Option) *Retry {
	r := &Retry{
		maxAttempts: nil,
		backoff:     nil,
		policy:      defaultPolicy,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func defaultPolicy(err error) bool {
	return true
}

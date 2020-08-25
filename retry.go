package retry

import (
	"context"
	"fmt"
	"time"
)

// Retry allows to retry a function until the amount of attempts is exhausted,
// using a backoff mechanism to wait between successive attempts.
type Retry struct {
	maxAttempts *int
	backoff     Backoff
}

// Backoff describes the backoff interface.
type Backoff interface {
	// Next returns the duration to wait before performing the next attempt.
	Next(attempt int) time.Duration
}

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
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := fn(ctx); err != nil {
			attempt++
			if r.maxAttempts != nil && attempt == *r.maxAttempts {
				return fmt.Errorf("all attempts have been exhausted, finished with error: %w", err)
			}
			if r.backoff != nil {
				time.Sleep(r.backoff.Next(attempt))
			}
			continue
		}
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

// New creates an instance of Retry.
// By defaults, there is no retry limit and no backoff.
func New(opts ...Option) *Retry {
	r := &Retry{
		maxAttempts: nil,
		backoff:     nil,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

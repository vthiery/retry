package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryDo(t *testing.T) {
	t.Parallel()

	retry := New(
		WithMaxAttempts(3),
	)

	attempt := 0
	assert.NoError(t, retry.Do(context.Background(), func(context.Context) error {
		attempt++

		return failFirstAttempts(3)(attempt)
	}))
	assert.Equal(t, 3, attempt)
}

var errFailAttempt = errors.New("fail this attempt")

func failFirstAttempts(numberOfFailures int) func(int) error {
	return func(attempt int) error {
		if attempt < numberOfFailures {
			return errFailAttempt
		}

		return nil
	}
}

func TestRetryDoNoAttempts(t *testing.T) {
	t.Parallel()

	retry := New(
		WithMaxAttempts(0),
	)

	assert.Error(t, retry.Do(context.Background(), func(context.Context) error {
		return nil
	}))
}

func TestRetryDoNoRetry(t *testing.T) {
	t.Parallel()

	retry := New(
		WithMaxAttempts(1),
	)

	attempt := 0
	assert.Error(t, retry.Do(context.Background(), func(context.Context) error {
		attempt++

		return errFailAttempt
	}))
	assert.Equal(t, 1, attempt)
}

func TestRetryExhaustedAttempts(t *testing.T) {
	t.Parallel()

	retry := New(
		WithMaxAttempts(3),
	)
	assert.False(t, retry.exhaustedAttempts(0))
	assert.False(t, retry.exhaustedAttempts(1))
	assert.False(t, retry.exhaustedAttempts(2))
	assert.True(t, retry.exhaustedAttempts(3))
	assert.True(t, retry.exhaustedAttempts(4))
	assert.True(t, retry.exhaustedAttempts(5))
}

func TestRetryExhaustedAttemptsSingle(t *testing.T) {
	t.Parallel()

	retry := New()
	assert.False(t, retry.exhaustedAttempts(100000))
}

func TestRetryDoWithCancelledContext(t *testing.T) {
	t.Parallel()

	retry := New()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()

	err := retry.Do(ctx, func(context.Context) error {
		time.Sleep(100 * time.Millisecond)

		return errFailAttempt
	})
	assert.True(t, errors.Is(err, context.Canceled))
}

type longBackoff struct{}

func (b longBackoff) Next(attempt int) time.Duration {
	return time.Hour
}

func TestRetryWaitBackoffTime(t *testing.T) {
	t.Parallel()

	retry := New(
		WithBackoff(&longBackoff{}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	assert.True(t, errors.Is(retry.waitBackoffTime(ctx, 1), context.Canceled))
}

func TestRetryWaitBackoffTimeNoBackoff(t *testing.T) {
	t.Parallel()

	retry := New()
	assert.NoError(t, retry.waitBackoffTime(context.Background(), 1))
}

func TestWait(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	assert.NoError(t, wait(ctx, 10*time.Millisecond))
}

func TestWaitWithCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	assert.True(t, errors.Is(wait(ctx, 1*time.Hour), context.Canceled))
}

func TestRetryDoWithBackoff(t *testing.T) {
	t.Parallel()

	retry := New(
		WithMaxAttempts(3),
		WithBackoff(NewConstantBackoff(2*time.Millisecond, 1*time.Millisecond)),
	)

	attempt := 0
	assert.NoError(t, retry.Do(context.Background(), func(context.Context) error {
		attempt++

		return failFirstAttempts(3)(attempt)
	}))
	assert.Equal(t, 3, attempt)
}

func TestRetryDoWithEarlyCancelledContext(t *testing.T) {
	t.Parallel()

	retry := New(
		WithMaxAttempts(10),
		WithBackoff(&longBackoff{}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	assert.True(t, errors.Is(retry.Do(ctx, func(context.Context) error {
		return errFailAttempt
	}), context.Canceled))
}

func TestNoAttemptsAllowedError(t *testing.T) {
	t.Parallel()

	err := NoAttemptsAllowedError{
		MaxAttempts: 0,
	}
	assert.Equal(t, "no attempts are allowed with max attempts set to 0", err.Error())
}

var errNonRetryable = errors.New("a non-retryable error")

func TestRetryDoWithPolicy(t *testing.T) {
	t.Parallel()

	retry := New(
		WithMaxAttempts(3),
		WithPolicy(
			func(err error) bool {
				return !errors.Is(err, errNonRetryable)
			},
		),
	)
	attempt := 0
	err := retry.Do(context.Background(), func(context.Context) error {
		attempt++

		return errNonRetryable
	})

	assert.True(t, errors.Is(err, errNonRetryable))
	assert.Equal(t, 1, attempt)
}

func BenchmarkRetryDo(b *testing.B) {
	maxAttempts := 5
	retry := New(
		WithMaxAttempts(maxAttempts),
		WithBackoff(
			NewExponentialBackoff(
				2*time.Millisecond,
				10*time.Millisecond,
				2*time.Millisecond,
			),
		),
	)

	operation := failFirstAttempts(maxAttempts - 1)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		attempt := 0
		assert.NoError(b, retry.Do(context.Background(), func(context.Context) error {
			attempt++

			return operation(attempt)
		}))
	}
}

package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryDo(t *testing.T) {
	r := New(
		WithMaxAttempts(3),
	)

	attempt := 0
	assert.NoError(t, r.Do(context.Background(), func(context.Context) error {
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
	r := New(
		WithMaxAttempts(0),
	)

	assert.Error(t, r.Do(context.Background(), func(context.Context) error {
		return nil
	}))
}

func TestRetryDoNoRetry(t *testing.T) {
	r := New(
		WithMaxAttempts(1),
	)

	attempt := 0
	assert.Error(t, r.Do(context.Background(), func(context.Context) error {
		attempt++
		return errFailAttempt
	}))
	assert.Equal(t, 1, attempt)
}

func TestRetryExhaustedAttempts(t *testing.T) {
	r := New(
		WithMaxAttempts(3),
	)
	assert.False(t, r.exhaustedAttempts(0))
	assert.False(t, r.exhaustedAttempts(1))
	assert.False(t, r.exhaustedAttempts(2))
	assert.True(t, r.exhaustedAttempts(3))
	assert.True(t, r.exhaustedAttempts(4))
	assert.True(t, r.exhaustedAttempts(5))
}

func TestRetryExhaustedAttemptsSingle(t *testing.T) {
	r := New()
	assert.False(t, r.exhaustedAttempts(100000))
}

type failBackoff struct {
	t *testing.T
}

func (b failBackoff) Next(attempt int) time.Duration {
	b.t.Fatalf("this backoff should never be called")
	return time.Second
}

func TestRetryDoWithCancelledContext(t *testing.T) {
	r := New(
		WithMaxAttempts(10),
		WithBackoff(&failBackoff{t}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := r.Do(ctx, func(context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return errFailAttempt
	})
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestRetryDoWithBackoff(t *testing.T) {
	r := New(
		WithMaxAttempts(3),
		WithBackoff(NewConstantBackoff(2*time.Millisecond, 1*time.Millisecond)),
	)

	attempt := 0
	assert.NoError(t, r.Do(context.Background(), func(context.Context) error {
		attempt++
		return failFirstAttempts(3)(attempt)
	}))
	assert.Equal(t, 3, attempt)
}

func TestNoAttemptsAllowedError(t *testing.T) {
	err := NoAttemptsAllowedError{
		MaxAttempts: 0,
	}
	assert.Equal(t, "no attempts are allowed with max attempts set to 0", err.Error())
}

func BenchmarkRetryDo(b *testing.B) {
	maxAttempts := 5
	r := New(
		WithMaxAttempts(maxAttempts),
		WithBackoff(
			NewExponentialBackoff(
				2*time.Millisecond,
				10*time.Millisecond,
				2*time.Millisecond,
			),
		),
	)
	fn := failFirstAttempts(maxAttempts - 1)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		attempt := 0
		assert.NoError(b, r.Do(context.Background(), func(context.Context) error {
			attempt++
			return fn(attempt)
		}))
	}
}

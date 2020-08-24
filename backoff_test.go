package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstantBackoffNext(t *testing.T) {
	wait := 100 * time.Millisecond
	maxJitter := 50 * time.Millisecond
	backoff := NewConstantBackoff(wait, maxJitter)

	assert.Equal(t, 0*time.Millisecond, backoff.Next(0))

	for i := 1; i < 100; i++ {
		n := backoff.Next(i)
		assert.True(t, wait <= n)
		assert.True(t, n <= wait+maxJitter)
	}
}

func TestConstantBackoffNextNoJitter(t *testing.T) {
	backoff := NewConstantBackoff(100*time.Millisecond, 0)

	assert.Equal(t, 0*time.Millisecond, backoff.Next(0))

	for i := 1; i < 100; i++ {
		assert.Equal(t, 100*time.Millisecond, backoff.Next(i))
	}
}

func TestExponentialBackoffNext(t *testing.T) {
	minWait := 2 * time.Millisecond
	maxWait := 10 * time.Millisecond
	maxJitter := 1 * time.Millisecond
	backoff := NewExponentialBackoff(minWait, maxWait, maxJitter)

	n := backoff.Next(0)
	assert.Equal(t, 0*time.Millisecond, n)

	n = backoff.Next(1)
	assert.True(t, minWait <= n)
	assert.True(t, n <= minWait+maxJitter)

	n = backoff.Next(2)
	assert.True(t, 2*minWait <= n)
	assert.True(t, n <= 2*minWait+maxJitter)

	n = backoff.Next(3)
	assert.True(t, 4*minWait <= n)
	assert.True(t, n <= 4*minWait+maxJitter)

	// Next times, the maximum wait time will be reached
	for i := 4; i < 100; i++ {
		assert.Equal(t, maxWait, backoff.Next(i))
	}
}

func TestExponentialBackoffNextNoJitter(t *testing.T) {
	minWait := 2 * time.Millisecond
	maxWait := 10 * time.Millisecond
	backoff := NewExponentialBackoff(minWait, maxWait, 0)

	assert.Equal(t, 0*time.Millisecond, backoff.Next(0))
	assert.Equal(t, minWait, backoff.Next(1))
	assert.Equal(t, 2*minWait, backoff.Next(2))
	assert.Equal(t, 4*minWait, backoff.Next(3))

	// Next times, the maximum wait time will be reached
	for i := 4; i < 100; i++ {
		assert.Equal(t, maxWait, backoff.Next(i))
	}
}

func TestJitter(t *testing.T) {
	assert.Equal(t, time.Duration(0), jitter(time.Duration(-42)))
	assert.Equal(t, time.Duration(0), jitter(time.Duration(0)))

	assert.True(t, jitter(time.Duration(42)) <= time.Duration(42))
}

func TestMinDuration(t *testing.T) {
	d1 := time.Duration(42)
	d2 := time.Duration(666)

	assert.Equal(t, d1, minDuration(d1, d2))
	assert.Equal(t, d1, minDuration(d2, d1))
}

func TestMaxDuration(t *testing.T) {
	d1 := time.Duration(42)
	d2 := time.Duration(666)

	assert.Equal(t, d2, maxDuration(d1, d2))
	assert.Equal(t, d2, maxDuration(d2, d1))
}

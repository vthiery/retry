package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstantBackoffNext(t *testing.T) {
	backoff := NewConstantBackoff(100*time.Millisecond, 50*time.Millisecond)

	assert.Equal(t, 0*time.Millisecond, backoff.Next(0))

	for i := 1; i < 100; i++ {
		n := backoff.Next(i)
		assert.True(t, 100*time.Millisecond <= n)
		assert.True(t, n <= 150*time.Millisecond)
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
	backoff := NewExponentialBackoff(2*time.Millisecond, 10*time.Millisecond, 1*time.Millisecond)

	n := backoff.Next(0)
	assert.Equal(t, 0*time.Millisecond, n)

	n = backoff.Next(1)
	assert.True(t, 2*time.Millisecond <= n)
	assert.True(t, n <= 3*time.Millisecond)

	n = backoff.Next(2)
	assert.True(t, 4*time.Millisecond <= n)
	assert.True(t, n <= 5*time.Millisecond)

	n = backoff.Next(3)
	assert.True(t, 8*time.Millisecond <= n)
	assert.True(t, n <= 9*time.Millisecond)

	// Next times, the maximum wait time will be reached
	for i := 4; i < 100; i++ {
		assert.Equal(t, 10*time.Millisecond, backoff.Next(i))
	}
}

func TestExponentialBackoffNextNoJitter(t *testing.T) {
	backoff := NewExponentialBackoff(2*time.Millisecond, 10*time.Millisecond, 0)

	assert.Equal(t, 0*time.Millisecond, backoff.Next(0))
	assert.Equal(t, 2*time.Millisecond, backoff.Next(1))
	assert.Equal(t, 4*time.Millisecond, backoff.Next(2))
	assert.Equal(t, 8*time.Millisecond, backoff.Next(3))

	// Next times, the maximum wait time will be reached
	for i := 4; i < 100; i++ {
		assert.Equal(t, 10*time.Millisecond, backoff.Next(i))
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

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

func TestExponentialBackoffNext(t *testing.T) {
	backoff := NewExponentialBackoff(2*time.Millisecond, 10*time.Millisecond, 2.0, 1*time.Millisecond)

	n := backoff.Next(0)
	assert.Equal(t, 0*time.Millisecond, n)

	n = backoff.Next(1)
	assert.True(t, 4*time.Millisecond <= n)
	assert.True(t, n <= 5*time.Millisecond)

	n = backoff.Next(2)
	assert.True(t, 6*time.Millisecond <= n)
	assert.True(t, n <= 7*time.Millisecond)

	// Next times, the maximum wait time will be reached
	for i := 3; i < 100; i++ {
		assert.Equal(t, 10*time.Millisecond, backoff.Next(i))
	}
}

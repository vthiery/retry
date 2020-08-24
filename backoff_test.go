package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstantBackoffNextTime(t *testing.T) {
	backoff := NewConstantBackoff(100*time.Millisecond, 50*time.Millisecond)

	assert.True(t, 0*time.Millisecond <= backoff.Next(0))
	assert.True(t, 100*time.Millisecond <= backoff.Next(1))
}

func TestExponentialBackoffNextTime(t *testing.T) {
	backoff := NewExponentialBackoff(2*time.Millisecond, 10*time.Millisecond, 2.0, 1*time.Millisecond)

	assert.True(t, 0*time.Millisecond <= backoff.Next(0))
	assert.True(t, 4*time.Millisecond <= backoff.Next(1))
}

func TestExponentialBackoffMaxTimeoutCrossed(t *testing.T) {
	backoff := NewExponentialBackoff(2*time.Millisecond, 9*time.Millisecond, 2.0, 1*time.Millisecond)

	assert.True(t, 9*time.Millisecond <= backoff.Next(3))
}

func TestExponentialBackoffMaxTimeoutReached(t *testing.T) {
	backoff := NewExponentialBackoff(2*time.Millisecond, 10*time.Millisecond, 2.0, 1*time.Millisecond)

	assert.True(t, 10*time.Millisecond <= backoff.Next(3))
}

func TestExponentialBackoffJitter(t *testing.T) {
	backoff := NewExponentialBackoff(2*time.Millisecond, 10*time.Millisecond, 2.0, 2*time.Millisecond)

	assert.True(t, 4*time.Millisecond <= backoff.Next(1))
}

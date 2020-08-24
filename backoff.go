package retry

import (
	"math"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type constantBackoff struct {
	backoffInterval       int64
	maximumJitterInterval int64
}

// NewConstantBackoff returns an instance of ConstantBackoff.
func NewConstantBackoff(backoffInterval, maximumJitterInterval time.Duration) Backoff {
	return &constantBackoff{
		backoffInterval:       int64(backoffInterval / time.Millisecond),
		maximumJitterInterval: int64(maximumJitterInterval / time.Millisecond),
	}
}

// Next returns next duration to wait before the next attempt.
func (b *constantBackoff) Next(attempt int) time.Duration {
	if attempt <= 0 {
		return 0 * time.Millisecond
	}
	delay := time.Duration(b.backoffInterval)
	jitter := time.Duration(rand.Int63n(b.maximumJitterInterval))
	return (delay + jitter) * time.Millisecond
}

type exponentialBackoff struct {
	exponentFactor        float64
	initialWait           float64
	maxWait               float64
	maximumJitterInterval int64
}

// Next returns next duration to wait before the next attempt.
func (b *exponentialBackoff) Next(attempt int) time.Duration {
	if attempt <= 0 {
		return 0 * time.Millisecond
	}
	delay := math.Min(b.initialWait+math.Pow(b.exponentFactor, float64(attempt)), b.maxWait)
	jitter := float64(rand.Int63n(b.maximumJitterInterval))
	return time.Duration(delay+jitter) * time.Millisecond
}

// NewExponentialBackoff returns an instance of ExponentialBackoff.
func NewExponentialBackoff(
	initialWait time.Duration,
	maxWait time.Duration,
	exponentFactor float64,
	maximumJitterInterval time.Duration,
) Backoff {
	return &exponentialBackoff{
		exponentFactor:        exponentFactor,
		initialWait:           float64(initialWait / time.Millisecond),
		maxWait:               float64(maxWait / time.Millisecond),
		maximumJitterInterval: int64(maximumJitterInterval / time.Millisecond),
	}
}

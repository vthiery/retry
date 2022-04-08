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
	wait      time.Duration
	maxJitter time.Duration
}

// NewConstantBackoff returns an instance of ConstantBackoff.
func NewConstantBackoff(wait, maxJitter time.Duration) Backoff {
	return &constantBackoff{
		wait:      maxDuration(wait, 0),
		maxJitter: maxDuration(maxJitter, 0),
	}
}

// Next returns next duration to wait before the next attempt.
func (b *constantBackoff) Next(attempt int) time.Duration {
	if attempt <= 0 {
		return 0 * time.Millisecond
	}

	return b.wait + jitter(b.maxJitter)
}

func jitter(maxJitter time.Duration) time.Duration {
	if maxJitter <= 0 {
		return 0
	}

	return time.Duration(rand.Int63n(int64(maxJitter))) // nolint:gosec
}

type exponentialBackoff struct {
	minWait   time.Duration
	maxWait   time.Duration
	maxJitter time.Duration
}

const limitDuration time.Duration = 1<<63 - 1

// Next returns next duration to wait before the next attempt.
func (b *exponentialBackoff) Next(attempt int) time.Duration {
	if attempt <= 0 {
		return 0 * time.Millisecond
	}
	// Make sure we don't overflow the time.Duration (int64)
	wait := float64(b.minWait) * math.Pow(2.0, float64(attempt-1)) // nolint
	if float64(limitDuration) < wait {
		return b.maxWait
	}

	return minDuration(time.Duration(wait)+jitter(b.maxJitter), b.maxWait)
}

func minDuration(d1, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d1
	}

	return d2
}

func maxDuration(d1, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d2
	}

	return d1
}

// NewExponentialBackoff returns an instance of ExponentialBackoff.
func NewExponentialBackoff(minWait, maxWait, maxJitter time.Duration) Backoff {
	return &exponentialBackoff{
		minWait:   maxDuration(minWait, 0),
		maxWait:   maxDuration(maxWait, 0),
		maxJitter: maxDuration(maxJitter, 0),
	}
}

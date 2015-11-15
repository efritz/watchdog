package watchdog

import (
	"time"
)

// BackOff is the interface to a back-off interval generator.
type BackOff interface {
	// Mark the next call to NextInterval as the "first" retry in a sequence.
	// If the generated intervals are dependent on the number of consecutive
	// (unsuccessful) retries, previous retries should be forgotten here.
	Reset()

	// Generate the next back-off interval.
	NextInterval() time.Duration
}

//
//

type zeroBackOff struct{}

func (b *zeroBackOff) Reset() {}

func (b *zeroBackOff) NextInterval() time.Duration {
	return 0 * time.Millisecond
}

// A back-off interval generator which always returns a zero interval.
func NewZeroBackOff() BackOff {
	return &zeroBackOff{}
}

//
//

type constantBackOff struct {
	interval time.Duration
}

func (b *constantBackOff) Reset() {}

func (b *constantBackOff) NextInterval() time.Duration {
	return b.interval
}

// A back-off interval generator which always returns the same interval.
func NewConstantBackOff(interval time.Duration) BackOff {
	return &constantBackOff{
		interval: interval,
	}
}

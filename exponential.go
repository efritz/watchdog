package watchdog

import (
	"math"
	"math/rand"
	"time"
)

const (
	DefaultMultiplier  = 1.25
	DefaultRandFactor  = 0.25
	DefaultMinInterval = 10 * time.Millisecond
	DefaultMaxInterval = 10 * time.Minute
)

type ExponentialBackOffConfig struct {
	Multiplier  float64
	RandFactor  float64
	MinInterval time.Duration
	MaxInterval time.Duration
}

// Create a new default ExponentialBackOffConfig.
func NewExponentialBackOffConfig() *ExponentialBackOffConfig {
	return &ExponentialBackOffConfig{
		Multiplier:  DefaultMultiplier,
		RandFactor:  DefaultRandFactor,
		MinInterval: DefaultMinInterval,
		MaxInterval: DefaultMaxInterval,
	}
}

//
//

type exponentialBackOff struct {
	attempts uint
	config   *ExponentialBackOffConfig
}

func (b *exponentialBackOff) Reset() {
	b.attempts = 0
}

func (b *exponentialBackOff) NextInterval() time.Duration {
	b.attempts += 1

	init := float64(b.config.MinInterval)
	base := b.config.Multiplier

	bInterval := init * math.Pow(base, float64(b.attempts))
	rInterval := time.Duration(randomNear(bInterval, b.config.RandFactor))

	if rInterval < b.config.MaxInterval {
		return rInterval
	} else {
		return b.config.MaxInterval
	}
}

func randomNear(value, ratio float64) float64 {
	min := value - (value * ratio)
	max := value + (value * ratio)

	return min + (max-min+1)*rand.Float64()
}

// A back-off interval generator which returns exponentially increasing
// intervals for each unsuccessful retry. The base interval is given by
// the function (MinInterval * Multiplier ^ n) where n is the number of
// previous failed attempts in the current sequence. The value returned
// is given by min(MaxInterval, base +/- (base * RandFactor)). A random
// factor of zero will make the generator deterministic.
func NewExponentialBackOff(config *ExponentialBackOffConfig) BackOff {
	b := &exponentialBackOff{
		attempts: 0,
		config:   config,
	}

	b.Reset()
	return b
}

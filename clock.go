package watchdog

import "time"

type (
	clock interface {
		After(duration time.Duration) <-chan time.Time
	}

	realClock struct{}
)

func (rc *realClock) After(duration time.Duration) <-chan time.Time {
	return time.After(duration)
}

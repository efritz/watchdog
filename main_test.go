package watchdog

import (
	"testing"
	"time"

	"github.com/efritz/backoff"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type WatchdogSuite struct{}

var _ = Suite(&WatchdogSuite{})

//
//

type MockRetry struct {
	f func() bool
}

func NewMockRetry(f func() bool) Retry {
	return &MockRetry{
		f: f,
	}
}

func (m *MockRetry) Retry() bool {
	return m.f()
}

//
//

type MockBackoff struct {
	f1 func()
	f2 func() time.Duration
}

	return &MockBackoff{
func NewMockBackoff(f1 func(), f2 func() time.Duration) backoff.Backoff {
		f1: f1,
		f2: f2,
	}
}

	return NewMockBackoff(func() {}, func() time.Duration {
func NewConstantBackoff(duration time.Duration) backoff.Backoff {
		return duration
	})
}

func (m *MockBackoff) Reset() {
	m.f1()
}

func (m *MockBackoff) NextInterval() time.Duration {
	return m.f2()
}

package watchdog

import (
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type WatchdogSuite struct{}

var _ = Suite(&WatchdogSuite{})

//
//

func testSequence(c *C, b BackOff, base time.Duration, durations []uint) {
	testRandomizedSequence(c, b, base, 0, durations)
}

func testRandomizedSequence(c *C, b BackOff, base time.Duration, ratio float64, durations []uint) {
	for _, duration := range durations {
		v := b.NextInterval()

		c.Assert(v >= time.Duration(float64(base)*float64(duration)*(1-ratio)), Equals, true)
		c.Assert(v <= time.Duration(float64(base)*float64(duration)*(1+ratio)), Equals, true)
	}
}

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

type MockBackOff struct {
	f1 func()
	f2 func() time.Duration
}

func NewMockBackOff(f1 func(), f2 func() time.Duration) BackOff {
	return &MockBackOff{
		f1: f1,
		f2: f2,
	}
}

func (m *MockBackOff) Reset() {
	m.f1()
}

func (m *MockBackOff) NextInterval() time.Duration {
	return m.f2()
}

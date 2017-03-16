package watchdog

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

type WatcherSuite struct{}

func (s *WatcherSuite) TestSuccess(t *testing.T) {
	var (
		attempts  = 0
		clockChan = make(chan time.Time)
		clock     = newMockClock(clockChan)
	)

	defer close(clockChan)

	watcher := newWatcherWithClock(
		RetryFunc(func() bool {
			attempts++
			return attempts >= 20
		}),
		&mockBackoff{},
		clock,
	)

	ch := watcher.Start()
	defer watcher.Stop()

	for i := 1; i < 20; i++ {
		clockChan <- time.Now()
	}

	<-ch
	Expect(attempts).To(Equal(20))
}

func (s *WatcherSuite) TestWatcherRespectsBackoff(t *testing.T) {
	var (
		attempts  = 0
		clockChan = make(chan time.Time)
		clock     = newMockClock(clockChan)
	)

	defer close(clockChan)

	watcher := newWatcherWithClock(
		RetryFunc(func() bool {
			attempts++
			return attempts >= 4
		}),
		&mockBackoff{},
		clock,
	)

	ch := watcher.Start()
	defer watcher.Stop()

	for i := 1; i < 4; i++ {
		clockChan <- time.Now()
	}

	<-ch
	Expect(attempts).To(Equal(4))
	Expect(clock.args).To(HaveLen(3))
}

func (s *WatcherSuite) TestStop(t *testing.T) {
	var (
		attempts  = 0
		clockChan = make(chan time.Time)
		clock     = newMockClock(clockChan)
		sync1     = make(chan struct{})
		sync2     = make(chan struct{})
	)

	defer close(sync1)
	defer close(sync2)
	defer close(clockChan)

	watcher := newWatcherWithClock(
		RetryFunc(func() bool {
			attempts++
			if attempts == 200 {
				sync1 <- struct{}{}
				<-sync2
			}
			return false
		}),
		&mockBackoff{},
		clock,
	)

	ch := watcher.Start()

	for i := 1; i < 200; i++ {
		clockChan <- time.Now()
	}

	<-sync1
	watcher.Stop()
	sync2 <- struct{}{}

	Expect(ch).To(BeClosedTimeout())
	Expect(attempts).To(Equal(200))
	Expect(clock.args).To(HaveLen(200))
}

func (s *WatcherSuite) TestCheck(t *testing.T) {
	var (
		attempts  = 0
		clockChan = make(chan time.Time)
		clock     = newMockClock(clockChan)
	)

	defer close(clockChan)

	watcher := newWatcherWithClock(
		RetryFunc(func() bool {
			attempts++
			return (attempts % 20) == 0
		}),
		&mockBackoff{},
		clock,
	)

	ch := watcher.Start()
	defer watcher.Stop()

	for i := 1; i < 20; i++ {
		clockChan <- time.Now()
	}

	<-ch
	Expect(attempts).To(Equal(20))
	watcher.Check()

	for i := 1; i < 20; i++ {
		clockChan <- time.Now()
	}

	<-ch
	Expect(attempts).To(Equal(40))
	watcher.Check()

	for i := 1; i < 20; i++ {
		clockChan <- time.Now()
	}

	<-ch
	Expect(attempts).To(Equal(60))
}

func (s *WatcherSuite) TestCheckDoesNotResetBackoffDuringWatch(t *testing.T) {
	var (
		attempts  = 0
		backoff   = &mockBackoff{}
		clockChan = make(chan time.Time)
		clock     = newMockClock(clockChan)
		sync1     = make(chan struct{})
		sync2     = make(chan struct{})
	)

	defer close(sync1)
	defer close(sync2)
	defer close(clockChan)

	watcher := newWatcherWithClock(
		RetryFunc(func() bool {
			attempts++
			if attempts == 200 {
				sync1 <- struct{}{}
				<-sync2
			}
			return false
		}),
		backoff,
		clock,
	)

	ch := watcher.Start()

	for i := 1; i < 200; i++ {
		watcher.Check()
		clockChan <- time.Now()
	}

	<-sync1
	watcher.Stop()
	sync2 <- struct{}{}

	Expect(ch).To(BeClosedTimeout())
	Expect(attempts).To(Equal(200))
	Expect(backoff.resets).To(Equal(1))
	Expect(clock.args).To(HaveLen(200))
}

func (s *WatcherSuite) TestCheckResetsBackoffAfterSuccess(t *testing.T) {
	var (
		attempts  = 0
		backoff   = &mockBackoff{}
		clockChan = make(chan time.Time)
		clock     = newMockClock(clockChan)
		sync1     = make(chan struct{})
		sync2     = make(chan struct{})
	)

	defer close(sync1)
	defer close(sync2)
	defer close(clockChan)

	watcher := newWatcherWithClock(
		RetryFunc(func() bool {
			attempts++
			return (attempts % 20) == 0
		}),
		backoff,
		clock,
	)

	ch := watcher.Start()
	defer watcher.Stop()

	for i := 1; i < 20; i++ {
		clockChan <- time.Now()
	}

	<-ch
	Expect(attempts).To(Equal(20))
	Expect(backoff.intervals).To(Equal(19))

	for j := 1; j <= 20; j++ {
		watcher.Check()

		for i := 1; i < 20; i++ {
			clockChan <- time.Now()
		}

		<-ch
		Expect(attempts).To(Equal((j + 1) * 20))
		Expect(backoff.intervals).To(Equal((j + 1) * 19))
	}
}

func (s *WatcherSuite) TestCheckDoesNotInterruptIntervalDuringWatch(t *testing.T) {
	var (
		attempts  = 0
		backoff   = &mockBackoff{}
		clockChan = make(chan time.Time)
		clock     = newMockClock(clockChan)
		sync1     = make(chan struct{})
		sync2     = make(chan struct{})
	)

	defer close(sync1)
	defer close(sync2)
	defer close(clockChan)

	watcher := newWatcherWithClock(
		RetryFunc(func() bool {
			attempts++
			return (attempts % 20) == 0
		}),
		backoff,
		clock,
	)

	ch := watcher.Start()
	defer watcher.Stop()

	for i := 1; i < 20; i++ {
		watcher.Check()
		clockChan <- time.Now()
	}

	<-ch
	Expect(attempts).To(Equal(20))
	Expect(backoff.resets).To(Equal(1))

	for j := 1; j <= 20; j++ {
		watcher.Check()

		for i := 1; i < 20; i++ {
			watcher.Check()
			clockChan <- time.Now()
		}

		<-ch
		Expect(attempts).To(Equal((j + 1) * 20))
		Expect(backoff.resets).To(Equal(j + 1))
	}
}

//
//
//

type mockClock struct {
	ch   <-chan time.Time
	args []time.Duration
}

func newMockClock(ch chan time.Time) *mockClock {
	return &mockClock{
		ch:   ch,
		args: []time.Duration{},
	}
}

func (m *mockClock) After(duration time.Duration) <-chan time.Time {
	ch := make(chan time.Time)
	m.args = append(m.args, duration)

	go func() {
		if t, ok := <-m.ch; ok {
			ch <- t
		}
	}()

	return ch
}

type mockBackoff struct {
	resets    int
	intervals int
}

func (m *mockBackoff) Reset() {
	m.resets++
}

func (m *mockBackoff) NextInterval() time.Duration {
	m.intervals++
	return 0
}

//
//
//

type BeClosedTimeoutMatcher struct{}

func (matcher *BeClosedTimeoutMatcher) Match(actual interface{}) (success bool, err error) {
	ch := actual.(<-chan struct{})

	select {
	case _, ok := <-ch:
		return !ok, nil
	case <-time.After(time.Second):
		return false, nil
	}
}

func BeClosedTimeout() *BeClosedTimeoutMatcher {
	return &BeClosedTimeoutMatcher{}
}

func (matcher *BeClosedTimeoutMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to be closed before timeout")
}

func (matcher *BeClosedTimeoutMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to be open")
}

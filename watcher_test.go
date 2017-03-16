package watchdog

import (
	"time"

	. "gopkg.in/check.v1"
)

func (s *WatchdogSuite) TestSuccess(c *C) {
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
	c.Assert(attempts, Equals, 20)
}

func (s *WatchdogSuite) TestWatcherRespectsBackoff(c *C) {
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
	c.Assert(attempts, Equals, 4)
	c.Assert(len(clock.args), Equals, 3)
}

func (s *WatchdogSuite) TestStop(c *C) {
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

	select {
	case _, ok := <-ch:
		c.Assert(ok, Equals, false)
	case <-time.After(time.Second):
		c.Errorf("expected success channel to be closed")
	}

	c.Assert(attempts, Equals, 200)
	c.Assert(len(clock.args), Equals, 200)
}

func (s *WatchdogSuite) TestCheck(c *C) {
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
	c.Assert(attempts, Equals, 20)
	watcher.Check()

	for i := 1; i < 20; i++ {
		clockChan <- time.Now()
	}

	<-ch
	c.Assert(attempts, Equals, 40)
	watcher.Check()

	for i := 1; i < 20; i++ {
		clockChan <- time.Now()
	}

	<-ch
	c.Assert(attempts, Equals, 60)
}

func (s *WatchdogSuite) TestCheckDoesNotResetBackoffDuringWatch(c *C) {
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

	select {
	case _, ok := <-ch:
		c.Assert(ok, Equals, false)
	case <-time.After(time.Second):
		c.Errorf("expected success channel to be closed")
	}

	c.Assert(attempts, Equals, 200)
	c.Assert(backoff.resets, Equals, 1)
	c.Assert(len(clock.args), Equals, 200)
}

func (s *WatchdogSuite) TestCheckResetsBackoffAfterSuccess(c *C) {
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
	c.Assert(attempts, Equals, 20)
	c.Assert(backoff.intervals, Equals, 19)

	for j := 1; j <= 20; j++ {
		watcher.Check()

		for i := 1; i < 20; i++ {
			clockChan <- time.Now()
		}

		<-ch
		c.Assert(attempts, Equals, (j+1)*20)
		c.Assert(backoff.intervals, Equals, (j+1)*19)
	}
}

func (s *WatchdogSuite) TestCheckDoesNotInterruptIntervalDuringWatch(c *C) {
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
	c.Assert(attempts, Equals, 20)
	c.Assert(backoff.resets, Equals, 1)

	for j := 1; j <= 20; j++ {
		watcher.Check()

		for i := 1; i < 20; i++ {
			watcher.Check()
			clockChan <- time.Now()
		}

		<-ch
		c.Assert(attempts, Equals, (j+1)*20)
		c.Assert(backoff.resets, Equals, j+1)
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

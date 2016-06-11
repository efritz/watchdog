package watchdog

import (
	"time"

	. "gopkg.in/check.v1"
)

func (s *WatchdogSuite) TestSuccess(c *C) {
	attempts := 0
	m := NewMockRetry(func() bool {
		attempts += 1
		return attempts >= 20
	})

	w := NewWatcher(m, NewConstantBackoff(0))
	w.Watch()
	<-w.Success

	c.Assert(attempts, Equals, 20)
}

func (s *WatchdogSuite) TestWatcherRespectsBackoff(c *C) {
	attempts := 0
	m := NewMockRetry(func() bool {
		attempts += 1
		return attempts >= 4
	})

	w := NewWatcher(m, NewConstantBackoff(time.Millisecond*200))
	w.Watch()

	select {
	case <-time.After(time.Millisecond * 500):
	case <-w.Success:
		c.Fatalf("Success happened too quickly.")
	}
}

func (s *WatchdogSuite) TestStop(c *C) {
	attempts := 0
	m := NewMockRetry(func() bool {
		attempts += 1
		return false
	})

	w := NewWatcher(m, NewConstantBackoff(0))
	w.Watch()

	<-time.After(50 * time.Millisecond)
	a1 := attempts

	<-time.After(50 * time.Millisecond)
	w.Stop()

	a2 := attempts
	<-time.After(50 * time.Millisecond)
	a3 := attempts

	c.Assert(a1, Not(Equals), a2)
	c.Assert(a2, Equals, a3)
}

func (s *WatchdogSuite) TestStopAfterSuccess(c *C) {
	attempts := 0
	m := NewMockRetry(func() bool {
		attempts += 1
		return true
	})

	w := NewWatcher(m, NewConstantBackoff(0))
	w.Watch()
	<-w.Success
	w.Stop()

	select {
	case w.ShouldRetry <- true:
		c.Fatalf("Watcher should not have accepted retry request.")
	default:
	}
}

func (s *WatchdogSuite) TestShouldRetry(c *C) {
	attempts := 0
	m := NewMockRetry(func() bool {
		attempts += 1
		return (attempts % 20) == 0
	})

	w := NewWatcher(m, NewConstantBackoff(0))
	w.Watch()
	<-w.Success
	c.Assert(attempts, Equals, 20)

	w.ShouldRetry <- true
	<-w.Success
	c.Assert(attempts, Equals, 40)

	w.ShouldRetry <- true
	<-w.Success
	c.Assert(attempts, Equals, 60)
}

func (s *WatchdogSuite) TestShouldRetryIgnored(c *C) {
	attempts := 0
	m := NewMockRetry(func() bool {
		attempts += 1
		return false
	})

	resets := 0
	b := NewMockBackoff(func() {
		resets += 1
	}, func() time.Duration {
		return 0 * time.Millisecond
	})

	w := NewWatcher(m, b)
	w.Watch()

	a1 := attempts
	<-time.After(50 * time.Millisecond)
	w.ShouldRetry <- true

	a2 := attempts
	<-time.After(50 * time.Millisecond)
	w.ShouldRetry <- true

	a3 := attempts
	<-time.After(50 * time.Millisecond)
	w.ShouldRetry <- true

	w.Stop()
	c.Assert(resets, Equals, 1)
	c.Assert(a1, Not(Equals), a2)
	c.Assert(a2, Not(Equals), a3)
}

func (s *WatchdogSuite) TestShouldRetryResetsBackoff(c *C) {
	attempts1 := 0
	m := NewMockRetry(func() bool {
		attempts1 += 1
		return (attempts1 % 20) == 0
	})

	attempts2 := 0
	b := NewMockBackoff(func() {
		attempts2 = 0
	}, func() time.Duration {
		attempts2 += 1
		return 0 * time.Millisecond
	})

	w := NewWatcher(m, b)
	w.Watch()
	<-w.Success
	c.Assert(attempts1, Equals, 20)
	c.Assert(attempts2, Equals, 19)

	w.ShouldRetry <- true
	<-w.Success
	c.Assert(attempts1, Equals, 40)
	c.Assert(attempts2, Equals, 19)
}

package watchdog

import (
	"time"

	. "gopkg.in/check.v1"
)

func (s *WatchdogSuite) TestSuccess(c *C) {
	attempts := 0
	m := NewMockRetry(func() bool {
		attempts++
		return attempts >= 20
	})

	w := NewWatcher(m, NewConstantBackoff(0))
	<-w.Start()

	c.Assert(attempts, Equals, 20)
}

func (s *WatchdogSuite) TestWatcherRespectsBackoff(c *C) {
	attempts := 0
	m := NewMockRetry(func() bool {
		attempts++
		return attempts >= 4
	})

	w := NewWatcher(m, NewConstantBackoff(time.Millisecond*200))
	ch := w.Start()

	select {
	case <-time.After(time.Millisecond * 500):
	case <-ch:
		c.Fatalf("Success happened too quickly.")
	}
}

func (s *WatchdogSuite) TestStop(c *C) {
	attempts := 0
	m := NewMockRetry(func() bool {
		attempts++
		return false
	})

	w := NewWatcher(m, NewConstantBackoff(0))
	w.Start()

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
		attempts++
		return true
	})

	w := NewWatcher(m, NewConstantBackoff(0))
	<-w.Start()
	w.Stop()
}

func (s *WatchdogSuite) TestStartAfterStop(c *C) {
	attempts := 0
	m := NewMockRetry(func() bool {
		attempts++
		return true
	})

	w := NewWatcher(m, NewConstantBackoff(0))
	<-w.Start()
	w.Stop()

	<-w.Start()
	w.Stop()
}

func (s *WatchdogSuite) TestCheck(c *C) {
	attempts := 0
	m := NewMockRetry(func() bool {
		attempts++
		return (attempts % 20) == 0
	})

	w := NewWatcher(m, NewConstantBackoff(0))
	ch := w.Start()
	<-ch
	c.Assert(attempts, Equals, 20)

	w.Check()
	<-ch
	c.Assert(attempts, Equals, 40)

	w.Check()
	<-ch
	c.Assert(attempts, Equals, 60)
}

func (s *WatchdogSuite) TestCheckDoesNotResetBackoffDuringWatch(c *C) {
	resets := 0
	attempts := 0

	m := NewMockRetry(func() bool {
		attempts++
		return false
	})

	b := NewMockBackoff(func() {
		resets++
	}, func() time.Duration {
		return 0 * time.Millisecond
	})

	w := NewWatcher(m, b)
	w.Start()

	a1 := attempts
	<-time.After(50 * time.Millisecond)
	w.Check()

	a2 := attempts
	<-time.After(50 * time.Millisecond)
	w.Check()

	a3 := attempts
	<-time.After(50 * time.Millisecond)
	w.Stop()

	c.Assert(resets, Equals, 1)
	c.Assert(a1, Not(Equals), a2)
	c.Assert(a2, Not(Equals), a3)
}

func (s *WatchdogSuite) TestCheckResetsBackoffAfterSuccess(c *C) {
	attempts := 0
	backoffs := 0

	m := NewMockRetry(func() bool {
		attempts++
		return (attempts % 20) == 0
	})

	b := NewMockBackoff(func() {
		backoffs = 0
	}, func() time.Duration {
		backoffs++
		return 0 * time.Millisecond
	})

	w := NewWatcher(m, b)
	ch := w.Start()
	<-ch
	c.Assert(attempts, Equals, 20)
	c.Assert(backoffs, Equals, 19)

	w.Check()
	<-ch
	c.Assert(attempts, Equals, 40)
	c.Assert(backoffs, Equals, 19)
}

func (s *WatchdogSuite) TestCheckDoesNotInterruptIntervalDuringWatch(c *C) {
	resets := 0
	checks := 0

	attempts := 0
	backoffs := 0
	stopChan := make(chan struct{})

	m := NewMockRetry(func() bool {
		attempts++
		if (attempts % 10) != 0 {
			return false
		}

		close(stopChan)
		return true
	})

	b := NewMockBackoff(func() {
		resets++
	}, func() time.Duration {
		backoffs++
		return 25 * time.Millisecond
	})

	w := NewWatcher(m, b)
	ch := w.Start()

	// Start a goroutine that hammers the check method on this watcher
	// while it's executing the invocation loop. If implemented incorrectly,
	// either the backoff should be called repeatedly, or the function should
	// never be called because of increasing wait intervals.

	go func() {
		for {
			select {
			case <-stopChan:
				return

			default:
				checks++
				w.Check()
			}
		}
	}()

	<-ch
	w.Stop()

	// Ensure we hit our attempt goals
	c.Assert(resets, Equals, 1)
	c.Assert(attempts, Equals, 10)
	c.Assert(backoffs, Equals, 9)

	// Make sure our check got through
	c.Assert(checks > 1000, Equals, true)
}

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

func (s *WatchdogSuite) TestCheckIgnored(c *C) {
	attempts := 0
	m := NewMockRetry(func() bool {
		attempts++
		return false
	})

	resets := 0
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

func (s *WatchdogSuite) TestCheckResetsBackoff(c *C) {
	attempts1 := 0
	m := NewMockRetry(func() bool {
		attempts1++
		return (attempts1 % 20) == 0
	})

	attempts2 := 0
	b := NewMockBackoff(func() {
		attempts2 = 0
	}, func() time.Duration {
		attempts2++
		return 0 * time.Millisecond
	})

	w := NewWatcher(m, b)
	ch := w.Start()
	<-ch
	c.Assert(attempts1, Equals, 20)
	c.Assert(attempts2, Equals, 19)

	w.Check()
	<-ch
	c.Assert(attempts1, Equals, 40)
	c.Assert(attempts2, Equals, 19)
}

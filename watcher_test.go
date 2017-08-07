package watchdog

import (
	"testing"
	"time"

	"github.com/efritz/glock"
	. "github.com/onsi/gomega"
)

type WatcherSuite struct{}

func (s *WatcherSuite) TestSuccess(t *testing.T) {
	var (
		attempts = 0
		clock    = glock.NewMockClock()
	)

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
		clock.BlockingAdvance(time.Second)
	}

	<-ch
	Expect(attempts).To(Equal(20))
}

func (s *WatcherSuite) TestWatcherRespectsBackoff(t *testing.T) {
	var (
		attempts = 0
		clock    = glock.NewMockClock()
	)

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
		clock.BlockingAdvance(time.Second)
	}

	<-ch
	Expect(attempts).To(Equal(4))
	Expect(clock.GetAfterArgs()).To(HaveLen(3))
}

func (s *WatcherSuite) TestStop(t *testing.T) {
	var (
		attempts = 0
		clock    = glock.NewMockClock()
		sync1    = make(chan struct{})
		sync2    = make(chan struct{})
	)

	defer close(sync1)
	defer close(sync2)

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
		clock.BlockingAdvance(time.Second)
	}

	<-sync1
	watcher.Stop()
	sync2 <- struct{}{}

	Eventually(ch).Should(BeClosed())
	Expect(attempts).To(Equal(200))
	Expect(clock.GetAfterArgs()).To(HaveLen(200))
}

func (s *WatcherSuite) TestCheck(t *testing.T) {
	var (
		attempts = 0
		clock    = glock.NewMockClock()
	)

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
		clock.BlockingAdvance(time.Second)
	}

	<-ch
	Expect(attempts).To(Equal(20))
	watcher.Check()

	for i := 1; i < 20; i++ {
		clock.BlockingAdvance(time.Second)
	}

	<-ch
	Expect(attempts).To(Equal(40))
	watcher.Check()

	for i := 1; i < 20; i++ {
		clock.BlockingAdvance(time.Second)
	}

	<-ch
	Expect(attempts).To(Equal(60))
}

func (s *WatcherSuite) TestCheckDoesNotResetBackoffDuringWatch(t *testing.T) {
	var (
		attempts = 0
		backoff  = &mockBackoff{}
		clock    = glock.NewMockClock()
		sync1    = make(chan struct{})
		sync2    = make(chan struct{})
	)

	defer close(sync1)
	defer close(sync2)

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
		clock.BlockingAdvance(time.Second)
	}

	<-sync1
	watcher.Stop()
	sync2 <- struct{}{}

	Eventually(ch).Should(BeClosed())
	Expect(attempts).To(Equal(200))
	Expect(backoff.resets).To(Equal(1))
	Expect(clock.GetAfterArgs()).To(HaveLen(200))
}

func (s *WatcherSuite) TestCheckResetsBackoffAfterSuccess(t *testing.T) {
	var (
		attempts = 0
		backoff  = &mockBackoff{}
		clock    = glock.NewMockClock()
		sync1    = make(chan struct{})
		sync2    = make(chan struct{})
	)

	defer close(sync1)
	defer close(sync2)

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
		clock.BlockingAdvance(time.Second)
	}

	<-ch
	Expect(attempts).To(Equal(20))
	Expect(backoff.intervals).To(Equal(19))

	for j := 1; j <= 20; j++ {
		watcher.Check()

		for i := 1; i < 20; i++ {
			clock.BlockingAdvance(time.Second)
		}

		<-ch
		Expect(attempts).To(Equal((j + 1) * 20))
		Expect(backoff.intervals).To(Equal((j + 1) * 19))
	}
}

func (s *WatcherSuite) TestCheckDoesNotInterruptIntervalDuringWatch(t *testing.T) {
	var (
		attempts = 0
		backoff  = &mockBackoff{}
		clock    = glock.NewMockClock()
		sync1    = make(chan struct{})
		sync2    = make(chan struct{})
	)

	defer close(sync1)
	defer close(sync2)

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
		clock.BlockingAdvance(time.Second)
	}

	<-ch
	Expect(attempts).To(Equal(20))
	Expect(backoff.resets).To(Equal(1))

	for j := 1; j <= 20; j++ {
		watcher.Check()

		for i := 1; i < 20; i++ {
			watcher.Check()
			clock.BlockingAdvance(time.Second)
		}

		<-ch
		Expect(attempts).To(Equal((j + 1) * 20))
		Expect(backoff.resets).To(Equal(j + 1))
	}
}

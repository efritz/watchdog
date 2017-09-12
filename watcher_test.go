package watchdog

import (
	"context"
	"testing"
	"time"

	"github.com/aphistic/sweet"
	"github.com/aphistic/sweet-junit"
	"github.com/efritz/glock"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) {
	RegisterFailHandler(sweet.GomegaFail)

	sweet.Run(m, func(s *sweet.S) {
		s.RegisterPlugin(junit.NewPlugin())

		s.AddSuite(&WatcherSuite{})
	})
}

type WatcherSuite struct{}

func (s *WatcherSuite) TestBlockUntilSuccess(t sweet.T) {
	attempts := 0
	f := RetryFunc(func() bool {
		attempts++
		return attempts == 2500
	})

	val := BlockUntilSuccess(f, &mockBackoff{}, context.Background())
	Expect(val).To(BeTrue())
	Expect(attempts).To(Equal(2500))
}

func (s *WatcherSuite) TestBlockUntilSuccessCanceled(t sweet.T) {
	var (
		f           = RetryFunc(func() bool { return false })
		ch          = make(chan bool)
		ctx, cancel = context.WithCancel(context.Background())
	)

	defer close(ch)

	go func() {
		ch <- BlockUntilSuccess(f, &mockBackoff{}, ctx)
	}()

	cancel()
	Eventually(ch).Should(Receive(BeFalse()))
}

func (s *WatcherSuite) TestSuccess(t sweet.T) {
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

func (s *WatcherSuite) TestWatcherRespectsBackoff(t sweet.T) {
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

func (s *WatcherSuite) TestStop(t sweet.T) {
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

func (s *WatcherSuite) TestCheck(t sweet.T) {
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

func (s *WatcherSuite) TestCheckDoesNotResetBackoffDuringWatch(t sweet.T) {
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

func (s *WatcherSuite) TestCheckResetsBackoffAfterSuccess(t sweet.T) {
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

func (s *WatcherSuite) TestCheckDoesNotInterruptIntervalDuringWatch(t sweet.T) {
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

//
//
//

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

package watchdog

import (
	"context"

	"github.com/efritz/backoff"
	"github.com/efritz/glock"
)

type (
	// Watcher invokes a Retry function until success.
	Watcher interface {
		// Start will *immediately* attempt to invoke the retry function. The
		// watcher will re-invoke the retry function on failure, with a delay
		// in between tries. On success, the watcher waits for a retry signal,
		// at which point the process repeats. The channel returned will receive
		// a value after the retry function returns true. The user should read a
		// value from this channel after a retry request signal is sent, as this
		// channel is un-buffered.
		Start() <-chan struct{}

		// Stop updates the watcher so that no future calls to the retry function
		// are attempted. This method must not be called twice.
		Stop()

		// Check requests watcher to re-invoke the retry function until success.
		// If the watcher is already in a retry cycle, then this function should
		// have no observable effect. This method must not be called after Stop.
		Check()
	}

	// Retry is the interface to something which are invoked until success.
	Retry interface {
		// Some critical action, which should return true on success.
		Retry() bool
	}

	// RetryFunc is a function that can be applied as a Retry.
	RetryFunc func() bool

	watcher struct {
		retry   Retry
		backoff backoff.Backoff
		clock   glock.Clock

		// The channel on which a quit signal can be sent. The watcher will
		// shutdown its goroutines after receiving a value on this channel.
		quit chan struct{}

		// The channel on which a restart request signal is sent. Once a
		// value is received on this channel, the watcher will execute the
		// retry function until success (or a quit signal is received). If
		// the watcher is already attempting to retry, any values received
		// on this channel will be ignored.
		restart chan struct{}
	}
)

// Retry will execute the RetryFunc.
func (f RetryFunc) Retry() bool {
	return f()
}

// BlockUntilSuccess creates a transient watcher that fires the given retry
// function until success. This method takes a context object which will,
// if canceled, will stop the watcher. Returns if the function succeeds and
// false if the method was canceled.
func BlockUntilSuccess(ctx context.Context, retry Retry, backoff backoff.Backoff) bool {
	watcher := NewWatcher(retry, backoff)
	defer watcher.Stop()

	select {
	case <-watcher.Start():
		return true
	case <-ctx.Done():
		return false
	}
}

// NewWatcher creates a new watcher with the given retry function and
// interval generator.
func NewWatcher(retry Retry, backoff backoff.Backoff) Watcher {
	return newWatcherWithClock(retry, backoff, glock.NewRealClock())
}

func newWatcherWithClock(retry Retry, backoff backoff.Backoff, clock glock.Clock) Watcher {
	return &watcher{
		retry:   retry,
		backoff: backoff,
		clock:   clock,
		quit:    make(chan struct{}),
		restart: make(chan struct{}),
	}
}

func (w *watcher) Start() <-chan struct{} {
	success := make(chan struct{})

	go func() {
		defer close(success)

		for {
			// Immediately try to invoke the function. If this fails, then
			// we'll reset our backoff interval generator and start the
			// invocation loop.

			if !w.retry.Retry() {
				w.backoff.Reset()

				if !w.invocationLoop() {
					return
				}
			}

			success <- struct{}{}

			select {
			case <-w.restart:
			case <-w.quit:
				return
			}
		}
	}()

	return (<-chan struct{})(success)
}

func (w *watcher) Stop() {
	close(w.quit)
}

func (w *watcher) Check() {
	w.restart <- struct{}{}
}

// Repeatedly invoke the retry function in a loop until either the
// function returns true or a signal is read from the quit channel.
// We sleep some time (respecting the backoff intervals) in between
// invocations. We'll read values from the restart channel to keep
// it clear, but we will not do anything special. Return true when
// the function returns because an invocation of the retry function
// was successful.
func (w *watcher) invocationLoop() bool {
	ch := make(chan struct{})
	defer close(ch)

	// Spawn a goroutine that will simply eat values off of the
	// restart channel so that we don't have to muck up the main
	// loop below. The follow goroutine will be cleaned up when
	// we close the channel ch created above.

	go func() {
		for {
			select {
			case <-ch:
				return
			case <-w.restart:
			}
		}
	}()

	for {
		interval := w.backoff.NextInterval()

		select {
		case <-w.clock.After(interval):
			if w.retry.Retry() {
				return true
			}

		case <-w.quit:
			return false
		}
	}
}

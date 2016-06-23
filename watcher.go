package watchdog

import (
	"time"

	"github.com/efritz/backoff"
)

// Backoff is the interface to a backoff interval generator. See the
// backoff dependency for details.
type Backoff backoff.Backoff

// Retry is the interface to something which are invoked until success.
type Retry interface {
	// Some critical action, which should return true on success.
	Retry() bool
}

// Watcher invokes a Retry function until success.
type Watcher struct {
	retry    Retry
	backoff  Backoff
	watching bool

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

// NewWatcher creates a new watcher with the given retry function and
// interval generator.
func NewWatcher(retry Retry, backoff Backoff) *Watcher {
	return &Watcher{
		retry:    retry,
		backoff:  backoff,
		watching: false,

		quit:    make(chan struct{}),
		restart: make(chan struct{}),
	}
}

// Start begins watching in a goroutine. The watcher will *immediately*
// attempt to invoke the retry function. The watcher will re-invoke the
// retry function on failure, with a delay in between tries. On success,
// the watcher waits for a retry signal, at which point the process
// repeats. The channel returned will receive a value after the retry
// function returns true. The user should read a value from this channel
// after a retry request signal is sent, as this channel is unbuffered.
func (w *Watcher) Start() <-chan struct{} {
	success := make(chan struct{})

	go func() {
		w.watching = true

		defer func() {
			w.watching = false
			close(success)
		}()

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

// Repeatedly invoke the retry function in a loop until either the
// function returns true or a signal is read from the quit channel.
// We sleep some time (respecting the backoff intervals) in between
// invocations. We'll read values from the restart channel to keep
// it clear, but we will not do anything special. Return true when
// the function halts because an invocation of the retry function
// was successful.
func (w *Watcher) invocationLoop() bool {
	for {
		interval := w.backoff.NextInterval()

		select {
		case <-time.After(interval):
			if w.retry.Retry() {
				return true
			}

		case <-w.restart:
		case <-w.quit:
			return false
		}
	}
}

// Stop kills the watcher routine so that no future calls to the
// retry function are attempted.
func (w *Watcher) Stop() {
	w.quit <- struct{}{}

}

// Restart will request the watcher to re-invoke the retry function
// until success. If the watcher is already in a retry cycle, then
// this function has no observable effect. This method does not do
// anything if the Stop method has been called.
func (w *Watcher) Restart() {
	if !w.watching {
		return
	}

	w.restart <- struct{}{}
}

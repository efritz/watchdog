package watchdog

import (
	"time"

	"github.com/efritz/backoff"
)

type Backoff backoff.Backoff

// Retry is the interface to something which are invoked until success.
type Retry interface {
	// Some critical action, which should return true on success.
	Retry() bool
}

// A Watcher invokes a Retry function until success.
type Watcher struct {
	retry   Retry
	backoff Backoff

	// The channel on which a quit signal can be sent. The watcher will
	// shutdown its goroutines after receiving a value on this channel.
	Quit chan bool

	// The channel on which a retry request signal can be sent. Once a
	// value is received on this channel, the watcher will execute the
	// retry function until success (or a quit signal is received). If
	// the watcher is already attempting to retry, any values received
	// on this channel will be ignored.
	ShouldRetry chan bool

	// The channel on which a successful retry signal will be sent. The
	// user should read a value from this channel after a retry request
	// signal is sent, as this channel is unbuffered.
	Success chan bool
}

// Create a new watcher with the given retry function and interval generator.
func NewWatcher(retry Retry, backoff Backoff) *Watcher {
	return &Watcher{
		retry:   retry,
		backoff: backoff,

		Quit:        make(chan bool),
		ShouldRetry: make(chan bool),
		Success:     make(chan bool),
	}
}

// Begin watching in a goroutine. The watcher will *immediately* attempt to
// invoke the retry function. The watcher will re-invoke the retry function
// on failure, with a delay in between tries. On success, the watcher waits
// for a retry signal, at which point the process repeats.
func (w *Watcher) Watch() {
	go func() {
		for {
			if !w.retry.Retry() {
				w.backoff.Reset()

			loop:
				for {
					interval := w.backoff.NextInterval()

					select {
					case <-time.After(interval):
						if w.retry.Retry() {
							break loop
						}

					case <-w.ShouldRetry:
					case <-w.Quit:
						return
					}
				}
			}

			w.Success <- true

			select {
			case <-w.ShouldRetry:
			case <-w.Quit:
				return
			}
		}
	}()
}

// Stop attempting to invoke the retry function.
func (w *Watcher) Stop() {
	w.Quit <- true
}

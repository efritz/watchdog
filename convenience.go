package watchdog

import (
	"time"

	"github.com/efritz/backoff"
)

// BlockUntilSuccess creates a transient watcher that fires the given
// retry function until success.
func BlockUntilSuccess(retry Retry, backoff backoff.Backoff) {
	BlockUntilSuccessOrQuit(retry, backoff, make(chan struct{}))
}

// BlockUntilSuccessTimeout creates a transient watcher that fires the
// given retry function until success or until the specified timeout
// elapses. This function returns true if the wrapped function returns
// true before quitting.
func BlockUntilSuccessOrTimeout(retry Retry, backoff backoff.Backoff, timeout time.Duration) bool {
	return BlockUntilSuccessOrQuit(retry, backoff, Signal(time.After(timeout)))
}

// BlockUntilSuccessorQuit creates a transient watcher that fires the
// given retry function until success or until a value is received on
// the given quit channel. This mefunctionthod returns true if the wrapped
// function returns true before quitting.
func BlockUntilSuccessOrQuit(retry Retry, backoff backoff.Backoff, quit <-chan struct{}) bool {
	w := NewWatcher(retry, backoff)
	defer w.Stop()

	select {
	case <-w.Start():
		return true
	case <-quit:
		return false
	}
}

// Signal returns a channel that closes after a value is received on the
// timeout channel.
func Signal(timeout <-chan time.Time) <-chan struct{} {
	ch := make(chan struct{})

	go func() {
		<-timeout
		close(ch)
	}()

	return ch
}

// QuitOrTimeout returns a channel that closes either after a timeout or
// once a signal is received on the given quit channel. This is meant to
// collapse two abort signals into a single one.
func QuitOrTimeout(duration time.Duration, quit <-chan struct{}) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)

		select {
		case <-quit:
		case <-time.After(duration):
		}
	}()

	return ch
}

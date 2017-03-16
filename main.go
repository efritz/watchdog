package watchdog

import "github.com/efritz/backoff"

// BlockUntilSuccess creates a watcher that fires until the given retry
// function returns a success, then disables the watcher. This function
// is synchronous.
func BlockUntilSuccess(retry Retry, backoff backoff.Backoff) {
	watcher := NewWatcher(retry, backoff)
	<-watcher.Start()
	watcher.Stop()
}

package watchdog

import (
	"fmt"
	"time"

	"github.com/efritz/backoff"
)

// BlockUntilSuccess creates a transient watcher that fires the given
// retry function until success.
func BlockUntilSuccess(retry Retry, backoff backoff.Backoff) {
	watcher := NewWatcher(retry, backoff)
	defer watcher.Stop()

	<-watcher.Start()
}

// BlockUntilSuccessTimeout creates a transient watcher that fires the
// given retry function until success or until the specified timeout
// elapses. An error is returned in the later case.
func BlockUntilSuccessTimeout(retry Retry, backoff backoff.Backoff, timeout time.Duration) error {
	watcher := NewWatcher(retry, backoff)
	defer watcher.Stop()

	select {
	case <-watcher.Start():
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout elapsed")
	}
}

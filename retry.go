package watchdog

type (
	// Retry is the interface to something which are invoked until success.
	Retry interface {
		// Some critical action, which should return true on success.
		Retry() bool
	}

	// RetryFunc is a function that can be applied as a Retry.
	RetryFunc func() bool
)

// Retry will execute the RetryFunc.
func (f RetryFunc) Retry() bool {
	return f()
}

# Watchdog

[![GoDoc](https://godoc.org/github.com/efritz/watchdog?status.svg)](https://godoc.org/github.com/efritz/watchdog)
[![Build Status](https://secure.travis-ci.org/efritz/watchdog.png)](http://travis-ci.org/efritz/watchdog)
[![codecov.io](http://codecov.io/github/efritz/watchdog/coverage.svg?branch=master)](http://codecov.io/github/efritz/watchdog?branch=master)

Go library for automatic service reconnection.

This library depends on the [backoff](https://github.com/efritz/backoff) library, which
defines structures for creating backoff interval generator.

## Example

First, you must define a type that conforms to the `Retry` interface. The interface has
a single method. In this example, the method attempts to dial a connection to a Riemann
server and returns true on success and false on failure.

```go
type RiemannService struct {
	conn    *riemann.Conn
	watcher *watchdog.Watcher
}

func (rs *RiemannService) Retry() bool {
	conn, err := riemann.Dial("localhost:5555")
	if err != nil {
		return false
	}

	rs.conn = conn
	return true
}
```

In order to ensure the application maintains a valid connection, create a watcher which
will attempt to reconnect to the service in another goroutine until success. In between
attempts the watcher will wait a configured interval (one second in this example).

```go
func NewRiemannService() *RiemannService {
	rs := &RiemannService{}

	// Create watcher on Riemann service
	w := NewWatcher(rs, NewConstantBackOff(time.Second))
	rs.watcher = w

	w.Watch()   // Connect initially
	<-w.Success // Block until success

	return rs
}
```

If you're using the service and detect a connection issue, you can alert the watcher
so that it attempts to reconnect. On successful reconnect, the watcher sends a value
on the `Success` channel. You can block on this read (to ensure a valid connection),
or you can read this value in a goroutine (but should not be ignored).

```go
func (rs *RiemannService) Write() {
	if conn == nil {
		rs.watcher.ShouldRetry <- true // Alert watcher of failure
		<-rs.watcher.Success           // Block until success
	}

	// Write omitted
}
```

Once you're done with the service, you should stop the watcher so that the background
reconnect goroutines can be shut down.

```go
func (rs *RiemannService) Close() {
	rs.watcher.Stop()
	rs.conn.Close()
}
```

## License

Copyright (c) 2015 Eric Fritz

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

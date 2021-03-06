# Watchdog

[![GoDoc](https://godoc.org/github.com/efritz/watchdog?status.svg)](https://godoc.org/github.com/efritz/watchdog)
[![Build Status](https://secure.travis-ci.org/efritz/watchdog.png)](http://travis-ci.org/efritz/watchdog)
[![Maintainability](https://api.codeclimate.com/v1/badges/9aab8d8dce9e96f2ab9a/maintainability)](https://codeclimate.com/github/efritz/watchdog/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/9aab8d8dce9e96f2ab9a/test_coverage)](https://codeclimate.com/github/efritz/watchdog/test_coverage)

Go library for automatic service reconnection.

This library depends on the [backoff](https://github.com/efritz/backoff) library, which
defines structures for creating backoff interval generator.

## Example

First, you must define a type that conforms to the `Retry` interface. The interface has
a single method. In this example, the method attempts to dial a connection to a Riemann
server and returns true on success and false on failure.

```go
type RiemannService struct {
	conn      *riemann.Conn
	watcher   *watchdog.Watcher
	connected <-chan struct{}
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
	w := NewWatcher(rs, NewConstantBackoff(time.Second))

	connected := w.Start() // Connect initially,
	<-connected            // Block until success

	rs.watcher = w
	rs.connected = connected
	return rs
}
```

If you're using the service and detect a connection issue, you can alert the watcher
so that it attempts to reconnect. On successful reconnect, the watcher sends a value
on a channel returned by the `Start` method. You can block on this read (to ensure a
valid connection), or you can read this value in a goroutine (but must be read to
prevent deadlocks within the watcher's background reconnect goroutine).

```go
func (rs *RiemannService) Write() {
	if conn == nil {
		rs.watcher.Check() // Alert watcher of failure
		<-rs.connected     // Block until success
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

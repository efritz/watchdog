# Watchdog

[![GoDoc](https://godoc.org/github.com/efritz/watchdog?status.svg)](https://godoc.org/github.com/efritz/watchdog)
[![Build Status](https://secure.travis-ci.org/efritz/watchdog.png)](http://travis-ci.org/efritz/watchdog)
[![codecov.io](http://codecov.io/github/efritz/watchdog/coverage.svg?branch=master)](http://codecov.io/github/efritz/watchdog?branch=master)

Go library for automatic service reconnection.

## Example

TODO

```go
type RiemannService struct {
	conn    *riemann.Connection
	watcher watchdog.Watcher
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

TODO

```go
func NewRiemannService() *RiemannService {
	rc := &RiemannService{}

	watcher := NewWatcher(rc, NewZeroBackOff())
	rc.watcher = watcher

	watcher.Watch()
	<-watcher.Success
	return rc
}
```

TODO

```go
func (rs *RiemannService) Write() {
	if conn == nil {
		rs.watcher.ShouldRetry <- true
		<-rs.watcher.Success
	}

	// Write omitted
}
```

TODO

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

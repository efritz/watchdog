package watchdog

import . "gopkg.in/check.v1"
import "time"

func (s *WatchdogSuite) TestBlockUntilSuccess(c *C) {
	attempts := 0

	BlockUntilSuccess(RetryFunc(func() bool {
		attempts++
		return attempts == 2500
	}), &mockBackoff{})

	c.Assert(attempts, Equals, 2500)
}

func (s *WatchdogSuite) TestBlockUntilSuccessTimeoutSuccess(c *C) {
	attempts := 0

	err := BlockUntilSuccessTimeout(RetryFunc(func() bool {
		attempts++
		return attempts == 2500
	}), &mockBackoff{}, time.Second)

	c.Assert(err, IsNil)
	c.Assert(attempts, Equals, 2500)
}

func (s *WatchdogSuite) TestBlockUntilSuccessTimeoutFailure(c *C) {
	err := BlockUntilSuccessTimeout(RetryFunc(func() bool {
		return false
	}), &mockBackoff{}, time.Millisecond*10)

	c.Assert(err, Not(IsNil))
}

package watchdog

import (
	. "gopkg.in/check.v1"
	"time"
)

func (s *WatchdogSuite) TestNonRandom(c *C) {
	conf := NewExponentialBackOffConfig()
	conf.RandFactor = 0
	conf.Multiplier = 2
	conf.MinInterval = time.Millisecond
	conf.MaxInterval = time.Minute

	b := NewExponentialBackOff(conf)

	testSequence(c, b, time.Millisecond, []uint{1, 2, 4, 8, 16, 32})
	b.Reset()
	testSequence(c, b, time.Millisecond, []uint{1, 2, 4, 8, 16, 32})
}

func (s *WatchdogSuite) TestMax(c *C) {
	conf := NewExponentialBackOffConfig()
	conf.RandFactor = 0
	conf.Multiplier = 2
	conf.MinInterval = time.Millisecond
	conf.MaxInterval = time.Millisecond * 4

	b := NewExponentialBackOff(conf)

	testSequence(c, b, time.Millisecond, []uint{1, 2, 4, 4, 4, 4})
	b.Reset()
	testSequence(c, b, time.Millisecond, []uint{1, 2, 4, 4, 4, 4})
}

func (s *WatchdogSuite) TestRandomized(c *C) {
	conf := NewExponentialBackOffConfig()
	conf.RandFactor = .25
	conf.Multiplier = 2
	conf.MinInterval = time.Millisecond
	conf.MaxInterval = time.Minute

	b := NewExponentialBackOff(conf)

	testRandomizedSequence(c, b, time.Millisecond, .25, []uint{1, 2, 4, 8, 16, 32})
	b.Reset()
	testRandomizedSequence(c, b, time.Millisecond, .25, []uint{1, 2, 4, 8, 16, 32})
}

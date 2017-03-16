package watchdog

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type WatchdogSuite struct{}

var _ = Suite(&WatchdogSuite{})

package watchdog

import (
	"testing"
	"time"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	sweet.T(func(s *sweet.S) {
		RegisterFailHandler(sweet.GomegaFail)

		s.RunSuite(t, &WatcherSuite{})
		s.RunSuite(t, &ConvenienceSuite{})
	})
}

//
//
//

type mockBackoff struct {
	resets    int
	intervals int
}

func (m *mockBackoff) Reset() {
	m.resets++
}

func (m *mockBackoff) NextInterval() time.Duration {
	m.intervals++
	return 0
}

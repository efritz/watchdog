package watchdog

import (
	"testing"
	"time"

	"github.com/aphistic/sweet"
	"github.com/aphistic/sweet-junit"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) {
	RegisterFailHandler(sweet.GomegaFail)

	sweet.Run(m, func(s *sweet.S) {
		s.RegisterPlugin(junit.NewPlugin())

		s.AddSuite(&WatcherSuite{})
		s.AddSuite(&ConvenienceSuite{})
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

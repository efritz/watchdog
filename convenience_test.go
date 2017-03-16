package watchdog

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

type ConvenienceSuite struct{}

func (s *ConvenienceSuite) TestBlockUntilSuccess(t *testing.T) {
	attempts := 0

	BlockUntilSuccess(RetryFunc(func() bool {
		attempts++
		return attempts == 2500
	}), &mockBackoff{})

	Expect(attempts).To(Equal(2500))
}

func (s *ConvenienceSuite) TestBlockUntilSuccessTimeoutSuccess(t *testing.T) {
	attempts := 0

	err := BlockUntilSuccessTimeout(RetryFunc(func() bool {
		attempts++
		return attempts == 2500
	}), &mockBackoff{}, time.Second)

	Expect(err).To(BeNil())
	Expect(attempts).To(Equal(2500))
}

func (s *ConvenienceSuite) TestBlockUntilSuccessTimeoutFailure(t *testing.T) {
	err := BlockUntilSuccessTimeout(RetryFunc(func() bool {
		return false
	}), &mockBackoff{}, time.Millisecond*10)

	Expect(err).NotTo(BeNil())
}

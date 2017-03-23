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

func (s *ConvenienceSuite) TestBlockUntilSuccessOrTimeoutSuccess(t *testing.T) {
	attempts := 0

	ok := BlockUntilSuccessOrTimeout(RetryFunc(func() bool {
		attempts++
		return attempts == 2500
	}), &mockBackoff{}, time.Second)

	Expect(ok).To(BeTrue())
	Expect(attempts).To(Equal(2500))
}

func (s *ConvenienceSuite) TestBlockUntilSuccessOrTimeoutFailure(t *testing.T) {
	ok := BlockUntilSuccessOrTimeout(RetryFunc(func() bool {
		return false
	}), &mockBackoff{}, time.Millisecond*10)

	Expect(ok).NotTo(BeTrue())
}

func (s *ConvenienceSuite) TestBlockUntilSuccessOrQuitSuccess(t *testing.T) {
	attempts := 0
	ch := make(chan struct{})
	defer close(ch)

	ok := BlockUntilSuccessOrQuit(RetryFunc(func() bool {
		attempts++
		return attempts == 2500
	}), &mockBackoff{}, ch)

	Expect(ok).To(BeTrue())
	Expect(attempts).To(Equal(2500))
}

func (s *ConvenienceSuite) TestBlockUntilSuccessOrQuitFailure(t *testing.T) {
	ch := make(chan struct{})
	go close(ch)

	ok := BlockUntilSuccessOrQuit(RetryFunc(func() bool {
		return false
	}), &mockBackoff{}, ch)

	Expect(ok).NotTo(BeTrue())
}

func (s *ConvenienceSuite) TestSignal(t *testing.T) {
	ch1 := make(chan time.Time)
	ch2 := Signal(ch1)

	Expect(ch2).ShouldNot(Receive())
	ch1 <- time.Now()
	Eventually(ch2).Should(BeClosed())
}

func (s *ConvenienceSuite) QuitOrTimeoutQuit(t *testing.T) {
	ch1 := make(chan struct{})
	ch2 := QuitOrTimeout(time.Millisecond, ch1)
	defer close(ch1)

	Expect(ch2).ShouldNot(Receive())
	Eventually(ch2).Should(BeClosed())
}

func (s *ConvenienceSuite) QuitOrTimeoutTimeout(t *testing.T) {
	ch1 := make(chan struct{})
	ch2 := QuitOrTimeout(time.Millisecond, ch1)

	Expect(ch2).ShouldNot(Receive())
	close(ch1)
	Expect(ch2).To(BeClosed())
}

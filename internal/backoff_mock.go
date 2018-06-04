// DO NOT EDIT
// Code generated automatically by github.com/efritz/go-mockgen
// $ go-mockgen github.com/efritz/backoff -d internal

package internal

import (
	backoff "github.com/efritz/backoff"
	time "time"
)

type MockBackoff struct {
	CloneFunc                  func() backoff.Backoff
	CloneFuncCallCount         int
	CloneFuncCallParams        []BackoffCloneParamSet
	NextIntervalFunc           func() time.Duration
	NextIntervalFuncCallCount  int
	NextIntervalFuncCallParams []BackoffNextIntervalParamSet
	ResetFunc                  func()
	ResetFuncCallCount         int
	ResetFuncCallParams        []BackoffResetParamSet
}
type BackoffCloneParamSet struct{}
type BackoffNextIntervalParamSet struct{}
type BackoffResetParamSet struct{}

var _ backoff.Backoff = NewMockBackoff()

func NewMockBackoff() *MockBackoff {
	m := &MockBackoff{}
	m.CloneFunc = m.defaultCloneFunc
	m.NextIntervalFunc = m.defaultNextIntervalFunc
	m.ResetFunc = m.defaultResetFunc
	return m
}
func (m *MockBackoff) Clone() backoff.Backoff {
	m.CloneFuncCallCount++
	m.CloneFuncCallParams = append(m.CloneFuncCallParams, BackoffCloneParamSet{})
	return m.CloneFunc()
}
func (m *MockBackoff) NextInterval() time.Duration {
	m.NextIntervalFuncCallCount++
	m.NextIntervalFuncCallParams = append(m.NextIntervalFuncCallParams, BackoffNextIntervalParamSet{})
	return m.NextIntervalFunc()
}
func (m *MockBackoff) Reset() {
	m.ResetFuncCallCount++
	m.ResetFuncCallParams = append(m.ResetFuncCallParams, BackoffResetParamSet{})
	m.ResetFunc()
}
func (m *MockBackoff) defaultCloneFunc() backoff.Backoff {
	return nil
}
func (m *MockBackoff) defaultNextIntervalFunc() time.Duration {
	return 0
}
func (m *MockBackoff) defaultResetFunc() {
	return
}

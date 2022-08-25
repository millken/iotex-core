// Code generated by MockGen. DO NOT EDIT.
// Source: ./consensus/fsm_context.go

// Package consensus is a generated GoMock package.
package consensus

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	go_fsm "github.com/iotexproject/go-fsm"
	zap "go.uber.org/zap"
)

// MockFSMContext is a mock of FSMContext interface.
type MockFSMContext struct {
	ctrl     *gomock.Controller
	recorder *MockFSMContextMockRecorder
}

// MockFSMContextMockRecorder is the mock recorder for MockFSMContext.
type MockFSMContextMockRecorder struct {
	mock *MockFSMContext
}

// NewMockFSMContext creates a new mock instance.
func NewMockFSMContext(ctrl *gomock.Controller) *MockFSMContext {
	mock := &MockFSMContext{ctrl: ctrl}
	mock.recorder = &MockFSMContextMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFSMContext) EXPECT() *MockFSMContextMockRecorder {
	return m.recorder
}

// AcceptBlockTTL mocks base method.
func (m *MockFSMContext) AcceptBlockTTL(arg0 uint64) time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AcceptBlockTTL", arg0)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// AcceptBlockTTL indicates an expected call of AcceptBlockTTL.
func (mr *MockFSMContextMockRecorder) AcceptBlockTTL(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AcceptBlockTTL", reflect.TypeOf((*MockFSMContext)(nil).AcceptBlockTTL), arg0)
}

// AcceptLockEndorsementTTL mocks base method.
func (m *MockFSMContext) AcceptLockEndorsementTTL(arg0 uint64) time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AcceptLockEndorsementTTL", arg0)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// AcceptLockEndorsementTTL indicates an expected call of AcceptLockEndorsementTTL.
func (mr *MockFSMContextMockRecorder) AcceptLockEndorsementTTL(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AcceptLockEndorsementTTL", reflect.TypeOf((*MockFSMContext)(nil).AcceptLockEndorsementTTL), arg0)
}

// AcceptProposalEndorsementTTL mocks base method.
func (m *MockFSMContext) AcceptProposalEndorsementTTL(arg0 uint64) time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AcceptProposalEndorsementTTL", arg0)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// AcceptProposalEndorsementTTL indicates an expected call of AcceptProposalEndorsementTTL.
func (mr *MockFSMContextMockRecorder) AcceptProposalEndorsementTTL(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AcceptProposalEndorsementTTL", reflect.TypeOf((*MockFSMContext)(nil).AcceptProposalEndorsementTTL), arg0)
}

// Activate mocks base method.
func (m *MockFSMContext) Activate(arg0 bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Activate", arg0)
}

// Activate indicates an expected call of Activate.
func (mr *MockFSMContextMockRecorder) Activate(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Activate", reflect.TypeOf((*MockFSMContext)(nil).Activate), arg0)
}

// Active mocks base method.
func (m *MockFSMContext) Active() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Active")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Active indicates an expected call of Active.
func (mr *MockFSMContextMockRecorder) Active() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Active", reflect.TypeOf((*MockFSMContext)(nil).Active))
}

// BlockInterval mocks base method.
func (m *MockFSMContext) BlockInterval(arg0 uint64) time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BlockInterval", arg0)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// BlockInterval indicates an expected call of BlockInterval.
func (mr *MockFSMContextMockRecorder) BlockInterval(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BlockInterval", reflect.TypeOf((*MockFSMContext)(nil).BlockInterval), arg0)
}

// Broadcast mocks base method.
func (m *MockFSMContext) Broadcast(arg0 interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Broadcast", arg0)
}

// Broadcast indicates an expected call of Broadcast.
func (mr *MockFSMContextMockRecorder) Broadcast(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Broadcast", reflect.TypeOf((*MockFSMContext)(nil).Broadcast), arg0)
}

// Commit mocks base method.
func (m *MockFSMContext) Commit(arg0 interface{}) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Commit indicates an expected call of Commit.
func (mr *MockFSMContextMockRecorder) Commit(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockFSMContext)(nil).Commit), arg0)
}

// CommitTTL mocks base method.
func (m *MockFSMContext) CommitTTL(arg0 uint64) time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CommitTTL", arg0)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// CommitTTL indicates an expected call of CommitTTL.
func (mr *MockFSMContextMockRecorder) CommitTTL(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommitTTL", reflect.TypeOf((*MockFSMContext)(nil).CommitTTL), arg0)
}

// Delay mocks base method.
func (m *MockFSMContext) Delay(arg0 uint64) time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delay", arg0)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// Delay indicates an expected call of Delay.
func (mr *MockFSMContextMockRecorder) Delay(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delay", reflect.TypeOf((*MockFSMContext)(nil).Delay), arg0)
}

// EventChanSize mocks base method.
func (m *MockFSMContext) EventChanSize() uint {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EventChanSize")
	ret0, _ := ret[0].(uint)
	return ret0
}

// EventChanSize indicates an expected call of EventChanSize.
func (mr *MockFSMContextMockRecorder) EventChanSize() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EventChanSize", reflect.TypeOf((*MockFSMContext)(nil).EventChanSize))
}

// Height mocks base method.
func (m *MockFSMContext) Height() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Height")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// Height indicates an expected call of Height.
func (mr *MockFSMContextMockRecorder) Height() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Height", reflect.TypeOf((*MockFSMContext)(nil).Height))
}

// IsDelegate mocks base method.
func (m *MockFSMContext) IsDelegate() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDelegate")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsDelegate indicates an expected call of IsDelegate.
func (mr *MockFSMContextMockRecorder) IsDelegate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDelegate", reflect.TypeOf((*MockFSMContext)(nil).IsDelegate))
}

// IsFutureEvent mocks base method.
func (m *MockFSMContext) IsFutureEvent(arg0 *ConsensusEvent) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsFutureEvent", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsFutureEvent indicates an expected call of IsFutureEvent.
func (mr *MockFSMContextMockRecorder) IsFutureEvent(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsFutureEvent", reflect.TypeOf((*MockFSMContext)(nil).IsFutureEvent), arg0)
}

// IsStaleEvent mocks base method.
func (m *MockFSMContext) IsStaleEvent(arg0 *ConsensusEvent) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsStaleEvent", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsStaleEvent indicates an expected call of IsStaleEvent.
func (mr *MockFSMContextMockRecorder) IsStaleEvent(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsStaleEvent", reflect.TypeOf((*MockFSMContext)(nil).IsStaleEvent), arg0)
}

// IsStaleUnmatchedEvent mocks base method.
func (m *MockFSMContext) IsStaleUnmatchedEvent(arg0 *ConsensusEvent) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsStaleUnmatchedEvent", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsStaleUnmatchedEvent indicates an expected call of IsStaleUnmatchedEvent.
func (mr *MockFSMContextMockRecorder) IsStaleUnmatchedEvent(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsStaleUnmatchedEvent", reflect.TypeOf((*MockFSMContext)(nil).IsStaleUnmatchedEvent), arg0)
}

// Logger mocks base method.
func (m *MockFSMContext) Logger() *zap.Logger {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Logger")
	ret0, _ := ret[0].(*zap.Logger)
	return ret0
}

// Logger indicates an expected call of Logger.
func (mr *MockFSMContextMockRecorder) Logger() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Logger", reflect.TypeOf((*MockFSMContext)(nil).Logger))
}

// NewBackdoorEvt mocks base method.
func (m *MockFSMContext) NewBackdoorEvt(arg0 go_fsm.State) *ConsensusEvent {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewBackdoorEvt", arg0)
	ret0, _ := ret[0].(*ConsensusEvent)
	return ret0
}

// NewBackdoorEvt indicates an expected call of NewBackdoorEvt.
func (mr *MockFSMContextMockRecorder) NewBackdoorEvt(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewBackdoorEvt", reflect.TypeOf((*MockFSMContext)(nil).NewBackdoorEvt), arg0)
}

// NewConsensusEvent mocks base method.
func (m *MockFSMContext) NewConsensusEvent(arg0 go_fsm.EventType, arg1 interface{}) *ConsensusEvent {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewConsensusEvent", arg0, arg1)
	ret0, _ := ret[0].(*ConsensusEvent)
	return ret0
}

// NewConsensusEvent indicates an expected call of NewConsensusEvent.
func (mr *MockFSMContextMockRecorder) NewConsensusEvent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewConsensusEvent", reflect.TypeOf((*MockFSMContext)(nil).NewConsensusEvent), arg0, arg1)
}

// NewLockEndorsement mocks base method.
func (m *MockFSMContext) NewLockEndorsement(arg0 interface{}) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewLockEndorsement", arg0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewLockEndorsement indicates an expected call of NewLockEndorsement.
func (mr *MockFSMContextMockRecorder) NewLockEndorsement(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewLockEndorsement", reflect.TypeOf((*MockFSMContext)(nil).NewLockEndorsement), arg0)
}

// NewPreCommitEndorsement mocks base method.
func (m *MockFSMContext) NewPreCommitEndorsement(arg0 interface{}) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewPreCommitEndorsement", arg0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewPreCommitEndorsement indicates an expected call of NewPreCommitEndorsement.
func (mr *MockFSMContextMockRecorder) NewPreCommitEndorsement(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewPreCommitEndorsement", reflect.TypeOf((*MockFSMContext)(nil).NewPreCommitEndorsement), arg0)
}

// NewProposalEndorsement mocks base method.
func (m *MockFSMContext) NewProposalEndorsement(arg0 interface{}) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewProposalEndorsement", arg0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewProposalEndorsement indicates an expected call of NewProposalEndorsement.
func (mr *MockFSMContextMockRecorder) NewProposalEndorsement(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewProposalEndorsement", reflect.TypeOf((*MockFSMContext)(nil).NewProposalEndorsement), arg0)
}

// PreCommitEndorsement mocks base method.
func (m *MockFSMContext) PreCommitEndorsement() interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PreCommitEndorsement")
	ret0, _ := ret[0].(interface{})
	return ret0
}

// PreCommitEndorsement indicates an expected call of PreCommitEndorsement.
func (mr *MockFSMContextMockRecorder) PreCommitEndorsement() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PreCommitEndorsement", reflect.TypeOf((*MockFSMContext)(nil).PreCommitEndorsement))
}

// Prepare mocks base method.
func (m *MockFSMContext) Prepare() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Prepare")
	ret0, _ := ret[0].(error)
	return ret0
}

// Prepare indicates an expected call of Prepare.
func (mr *MockFSMContextMockRecorder) Prepare() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Prepare", reflect.TypeOf((*MockFSMContext)(nil).Prepare))
}

// Proposal mocks base method.
func (m *MockFSMContext) Proposal() (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Proposal")
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Proposal indicates an expected call of Proposal.
func (mr *MockFSMContextMockRecorder) Proposal() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Proposal", reflect.TypeOf((*MockFSMContext)(nil).Proposal))
}

// UnmatchedEventInterval mocks base method.
func (m *MockFSMContext) UnmatchedEventInterval(arg0 uint64) time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnmatchedEventInterval", arg0)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// UnmatchedEventInterval indicates an expected call of UnmatchedEventInterval.
func (mr *MockFSMContextMockRecorder) UnmatchedEventInterval(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnmatchedEventInterval", reflect.TypeOf((*MockFSMContext)(nil).UnmatchedEventInterval), arg0)
}

// UnmatchedEventTTL mocks base method.
func (m *MockFSMContext) UnmatchedEventTTL(arg0 uint64) time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnmatchedEventTTL", arg0)
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// UnmatchedEventTTL indicates an expected call of UnmatchedEventTTL.
func (mr *MockFSMContextMockRecorder) UnmatchedEventTTL(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnmatchedEventTTL", reflect.TypeOf((*MockFSMContext)(nil).UnmatchedEventTTL), arg0)
}

// WaitUntilRoundStart mocks base method.
func (m *MockFSMContext) WaitUntilRoundStart() time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitUntilRoundStart")
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// WaitUntilRoundStart indicates an expected call of WaitUntilRoundStart.
func (mr *MockFSMContextMockRecorder) WaitUntilRoundStart() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitUntilRoundStart", reflect.TypeOf((*MockFSMContext)(nil).WaitUntilRoundStart))
}

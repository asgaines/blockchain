// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/asgaines/blockchain/mining (interfaces: Miner)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	chain "github.com/asgaines/blockchain/chain"
	blockchain "github.com/asgaines/blockchain/protogo/blockchain"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockMiner is a mock of Miner interface
type MockMiner struct {
	ctrl     *gomock.Controller
	recorder *MockMinerMockRecorder
}

// MockMinerMockRecorder is the mock recorder for MockMiner
type MockMinerMockRecorder struct {
	mock *MockMiner
}

// NewMockMiner creates a new mock instance
func NewMockMiner(ctrl *gomock.Controller) *MockMiner {
	mock := &MockMiner{ctrl: ctrl}
	mock.recorder = &MockMinerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMiner) EXPECT() *MockMinerMockRecorder {
	return m.recorder
}

// AddTx mocks base method
func (m *MockMiner) AddTx(arg0 *blockchain.Tx) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddTx", arg0)
}

// AddTx indicates an expected call of AddTx
func (mr *MockMinerMockRecorder) AddTx(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTx", reflect.TypeOf((*MockMiner)(nil).AddTx), arg0)
}

// ClearTxs mocks base method
func (m *MockMiner) ClearTxs() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ClearTxs")
}

// ClearTxs indicates an expected call of ClearTxs
func (mr *MockMinerMockRecorder) ClearTxs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClearTxs", reflect.TypeOf((*MockMiner)(nil).ClearTxs))
}

// Mine mocks base method
func (m *MockMiner) Mine(arg0 context.Context, arg1 chan<- *chain.Block) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Mine", arg0, arg1)
}

// Mine indicates an expected call of Mine
func (mr *MockMinerMockRecorder) Mine(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Mine", reflect.TypeOf((*MockMiner)(nil).Mine), arg0, arg1)
}

// SetTarget mocks base method
func (m *MockMiner) SetTarget(arg0 float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTarget", arg0)
}

// SetTarget indicates an expected call of SetTarget
func (mr *MockMinerMockRecorder) SetTarget(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTarget", reflect.TypeOf((*MockMiner)(nil).SetTarget), arg0)
}

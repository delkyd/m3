// Automatically generated by MockGen. DO NOT EDIT!
// Source: session.go

// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package mocks

import (
	time "time"

	gomock "github.com/golang/mock/gomock"
	"github.com/m3db/m3db/interfaces/m3db"
	time0 "github.com/m3db/m3db/x/time"
)

// Mock of clientSession interface
type MockclientSession struct {
	ctrl     *gomock.Controller
	recorder *_MockclientSessionRecorder
}

// Recorder for MockclientSession (not exported)
type _MockclientSessionRecorder struct {
	mock *MockclientSession
}

func NewMockclientSession(ctrl *gomock.Controller) *MockclientSession {
	mock := &MockclientSession{ctrl: ctrl}
	mock.recorder = &_MockclientSessionRecorder{mock}
	return mock
}

func (_m *MockclientSession) EXPECT() *_MockclientSessionRecorder {
	return _m.recorder
}

func (_m *MockclientSession) Write(id string, t time.Time, value float64, unit time0.Unit, annotation []byte) error {
	ret := _m.ctrl.Call(_m, "Write", id, t, value, unit, annotation)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockclientSessionRecorder) Write(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Write", arg0, arg1, arg2, arg3, arg4)
}

func (_m *MockclientSession) Fetch(id string, startInclusive time.Time, endExclusive time.Time) (m3db.SeriesIterator, error) {
	ret := _m.ctrl.Call(_m, "Fetch", id, startInclusive, endExclusive)
	ret0, _ := ret[0].(m3db.SeriesIterator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockclientSessionRecorder) Fetch(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Fetch", arg0, arg1, arg2)
}

func (_m *MockclientSession) FetchAll(ids []string, startInclusive time.Time, endExclusive time.Time) (m3db.SeriesIterators, error) {
	ret := _m.ctrl.Call(_m, "FetchAll", ids, startInclusive, endExclusive)
	ret0, _ := ret[0].(m3db.SeriesIterators)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockclientSessionRecorder) FetchAll(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "FetchAll", arg0, arg1, arg2)
}

func (_m *MockclientSession) Close() error {
	ret := _m.ctrl.Call(_m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockclientSessionRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockclientSession) Open() error {
	ret := _m.ctrl.Call(_m, "Open")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockclientSessionRecorder) Open() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Open")
}
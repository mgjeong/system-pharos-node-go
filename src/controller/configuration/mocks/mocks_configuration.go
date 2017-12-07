// Automatically generated by MockGen. DO NOT EDIT!
// Source: configuration.go

package mock_configuration

import (
        gomock "github.com/golang/mock/gomock"
)

// Mock of Command interface
type MockCommand struct {
        ctrl     *gomock.Controller
        recorder *_MockCommandRecorder
}

// Recorder for MockCommand (not exported)
type _MockCommandRecorder struct {
        mock *MockCommand
}

func NewMockCommand(ctrl *gomock.Controller) *MockCommand {
        mock := &MockCommand{ctrl: ctrl}
        mock.recorder = &_MockCommandRecorder{mock}
        return mock
}

func (_m *MockCommand) EXPECT() *_MockCommandRecorder {
        return _m.recorder
}

func (_m *MockCommand) GetConfiguration() (map[string]interface{}, error) {
        ret := _m.ctrl.Call(_m, "GetConfiguration")
        ret0, _ := ret[0].(map[string]interface{})
        ret1, _ := ret[1].(error)
        return ret0, ret1
}

func (_mr *_MockCommandRecorder) GetConfiguration() *gomock.Call {
        return _mr.mock.ctrl.RecordCall(_mr.mock, "GetConfiguration")
}

func (_m *MockCommand) SetConfiguration(_param0 map[string]interface{}) error {
        ret := _m.ctrl.Call(_m, "SetConfiguration", _param0)
        ret0, _ := ret[0].(error)
        return ret0
}

func (_mr *_MockCommandRecorder) SetConfiguration(arg0 interface{}) *gomock.Call {
        return _mr.mock.ctrl.RecordCall(_mr.mock, "SetConfiguration", arg0)
}

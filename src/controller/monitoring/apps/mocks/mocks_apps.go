/*******************************************************************************
 * Copyright 2018 Samsung Electronics All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *******************************************************************************/
// Code generated by MockGen. DO NOT EDIT.
// Source: src/controller/monitoring/apps/apps.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockCommand is a mock of Command interface
type MockCommand struct {
	ctrl     *gomock.Controller
	recorder *MockCommandMockRecorder
}

// MockCommandMockRecorder is the mock recorder for MockCommand
type MockCommandMockRecorder struct {
	mock *MockCommand
}

// NewMockCommand creates a new mock instance
func NewMockCommand(ctrl *gomock.Controller) *MockCommand {
	mock := &MockCommand{ctrl: ctrl}
	mock.recorder = &MockCommandMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCommand) EXPECT() *MockCommandMockRecorder {
	return m.recorder
}

// EnableEventMonitoring mocks base method
func (m *MockCommand) EnableEventMonitoring(appId, path string) error {
	ret := m.ctrl.Call(m, "EnableEventMonitoring", appId, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// EnableEventMonitoring indicates an expected call of EnableEventMonitoring
func (mr *MockCommandMockRecorder) EnableEventMonitoring(appId, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnableEventMonitoring", reflect.TypeOf((*MockCommand)(nil).EnableEventMonitoring), appId, path)
}

// DisableEventMonitoring mocks base method
func (m *MockCommand) DisableEventMonitoring(appId, path string) error {
	ret := m.ctrl.Call(m, "DisableEventMonitoring", appId, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// DisableEventMonitoring indicates an expected call of DisableEventMonitoring
func (mr *MockCommandMockRecorder) DisableEventMonitoring(appId, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DisableEventMonitoring", reflect.TypeOf((*MockCommand)(nil).DisableEventMonitoring), appId, path)
}

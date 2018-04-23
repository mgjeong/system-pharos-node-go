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
// Source: apps.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	http "net/http"
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

// Handle mocks base method
func (m *MockCommand) Handle(w http.ResponseWriter, req *http.Request) {
	m.ctrl.Call(m, "Handle", w, req)
}

// Handle indicates an expected call of Handle
func (mr *MockCommandMockRecorder) Handle(w, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Handle", reflect.TypeOf((*MockCommand)(nil).Handle), w, req)
}

// MockapiInnerCommand is a mock of apiInnerCommand interface
type MockapiInnerCommand struct {
	ctrl     *gomock.Controller
	recorder *MockapiInnerCommandMockRecorder
}

// MockapiInnerCommandMockRecorder is the mock recorder for MockapiInnerCommand
type MockapiInnerCommandMockRecorder struct {
	mock *MockapiInnerCommand
}

// NewMockapiInnerCommand creates a new mock instance
func NewMockapiInnerCommand(ctrl *gomock.Controller) *MockapiInnerCommand {
	mock := &MockapiInnerCommand{ctrl: ctrl}
	mock.recorder = &MockapiInnerCommandMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockapiInnerCommand) EXPECT() *MockapiInnerCommandMockRecorder {
	return m.recorder
}

// subscribeEvent mocks base method
func (m *MockapiInnerCommand) subscribeEvent(w http.ResponseWriter, req *http.Request) {
	m.ctrl.Call(m, "subscribeEvent", w, req)
}

// subscribeEvent indicates an expected call of subscribeEvent
func (mr *MockapiInnerCommandMockRecorder) subscribeEvent(w, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "subscribeEvent", reflect.TypeOf((*MockapiInnerCommand)(nil).subscribeEvent), w, req)
}

// unsubscribeEvent mocks base method
func (m *MockapiInnerCommand) unsubscribeEvent(w http.ResponseWriter, req *http.Request) {
	m.ctrl.Call(m, "unsubscribeEvent", w, req)
}

// unsubscribeEvent indicates an expected call of unsubscribeEvent
func (mr *MockapiInnerCommandMockRecorder) unsubscribeEvent(w, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "unsubscribeEvent", reflect.TypeOf((*MockapiInnerCommand)(nil).unsubscribeEvent), w, req)
}

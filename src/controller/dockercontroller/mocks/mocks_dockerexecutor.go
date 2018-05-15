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
// Source: src/controller/dockercontroller/dockerexecutor.go

// Package mock_dockercontroller is a generated GoMock package.
package mock_dockercontroller

import (
	"controller/dockercontroller"
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

// Create mocks base method
func (m *MockCommand) Create(id, path string) error {
	ret := m.ctrl.Call(m, "Create", id, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockCommandMockRecorder) Create(id, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockCommand)(nil).Create), id, path)
}

// Up mocks base method
func (m *MockCommand) Up(id, path string, services ...string) error {
	varargs := []interface{}{id, path}
	for _, a := range services {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Up", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Up indicates an expected call of Up
func (mr *MockCommandMockRecorder) Up(id, path interface{}, services ...interface{}) *gomock.Call {
	varargs := append([]interface{}{id, path}, services...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Up", reflect.TypeOf((*MockCommand)(nil).Up), varargs...)
}

// Down mocks base method
func (m *MockCommand) Down(id, path string) error {
	ret := m.ctrl.Call(m, "Down", id, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// Down indicates an expected call of Down
func (mr *MockCommandMockRecorder) Down(id, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Down", reflect.TypeOf((*MockCommand)(nil).Down), id, path)
}

// DownWithRemoveImages mocks base method
func (m *MockCommand) DownWithRemoveImages(id, path string) error {
	ret := m.ctrl.Call(m, "DownWithRemoveImages", id, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// DownWithRemoveImages indicates an expected call of DownWithRemoveImages
func (mr *MockCommandMockRecorder) DownWithRemoveImages(id, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownWithRemoveImages", reflect.TypeOf((*MockCommand)(nil).DownWithRemoveImages), id, path)
}

// Start mocks base method
func (m *MockCommand) Start(id, path string) error {
	ret := m.ctrl.Call(m, "Start", id, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start
func (mr *MockCommandMockRecorder) Start(id, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockCommand)(nil).Start), id, path)
}

// Stop mocks base method
func (m *MockCommand) Stop(id, path string) error {
	ret := m.ctrl.Call(m, "Stop", id, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop
func (mr *MockCommandMockRecorder) Stop(id, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockCommand)(nil).Stop), id, path)
}

// Pause mocks base method
func (m *MockCommand) Pause(id, path string) error {
	ret := m.ctrl.Call(m, "Pause", id, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// Pause indicates an expected call of Pause
func (mr *MockCommandMockRecorder) Pause(id, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pause", reflect.TypeOf((*MockCommand)(nil).Pause), id, path)
}

// Unpause mocks base method
func (m *MockCommand) Unpause(id, path string) error {
	ret := m.ctrl.Call(m, "Unpause", id, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unpause indicates an expected call of Unpause
func (mr *MockCommandMockRecorder) Unpause(id, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unpause", reflect.TypeOf((*MockCommand)(nil).Unpause), id, path)
}

// Pull mocks base method
func (m *MockCommand) Pull(id, path string, services ...string) error {
	varargs := []interface{}{id, path}
	for _, a := range services {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Pull", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Pull indicates an expected call of Pull
func (mr *MockCommandMockRecorder) Pull(id, path interface{}, services ...interface{}) *gomock.Call {
	varargs := append([]interface{}{id, path}, services...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pull", reflect.TypeOf((*MockCommand)(nil).Pull), varargs...)
}

// Ps mocks base method
func (m *MockCommand) Ps(id, path string, args ...string) ([]map[string]string, error) {
	varargs := []interface{}{id, path}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Ps", varargs...)
	ret0, _ := ret[0].([]map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Ps indicates an expected call of Ps
func (mr *MockCommandMockRecorder) Ps(id, path interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{id, path}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ps", reflect.TypeOf((*MockCommand)(nil).Ps), varargs...)
}

// GetAppStats mocks base method
func (m *MockCommand) GetAppStats(id, path string) ([]map[string]interface{}, error) {
	ret := m.ctrl.Call(m, "GetAppStats", id, path)
	ret0, _ := ret[0].([]map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAppStats indicates an expected call of GetAppStats
func (mr *MockCommandMockRecorder) GetAppStats(id, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAppStats", reflect.TypeOf((*MockCommand)(nil).GetAppStats), id, path)
}

// GetContainerConfigByName mocks base method
func (m *MockCommand) GetContainerConfigByName(containerName string) (map[string]interface{}, error) {
	ret := m.ctrl.Call(m, "GetContainerConfigByName", containerName)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContainerConfigByName indicates an expected call of GetContainerConfigByName
func (mr *MockCommandMockRecorder) GetContainerConfigByName(containerName interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContainerConfigByName", reflect.TypeOf((*MockCommand)(nil).GetContainerConfigByName), containerName)
}

// GetImageDigestByName mocks base method
func (m *MockCommand) GetImageDigestByName(imageName string) (string, error) {
	ret := m.ctrl.Call(m, "GetImageDigestByName", imageName)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetImageDigestByName indicates an expected call of GetImageDigestByName
func (mr *MockCommandMockRecorder) GetImageDigestByName(imageName interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetImageDigestByName", reflect.TypeOf((*MockCommand)(nil).GetImageDigestByName), imageName)
}

// GetImageIDByRepoDigest mocks base method
func (m *MockCommand) GetImageIDByRepoDigest(imageName string) (string, error) {
	ret := m.ctrl.Call(m, "GetImageIDByRepoDigest", imageName)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetImageIDByRepoDigest indicates an expected call of GetImageIDByRepoDigest
func (mr *MockCommandMockRecorder) GetImageIDByRepoDigest(imageName interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetImageIDByRepoDigest", reflect.TypeOf((*MockCommand)(nil).GetImageIDByRepoDigest), imageName)
}

// ImagePull mocks base method
func (m *MockCommand) ImagePull(image string) error {
	ret := m.ctrl.Call(m, "ImagePull", image)
	ret0, _ := ret[0].(error)
	return ret0
}

// ImagePull indicates an expected call of ImagePull
func (mr *MockCommandMockRecorder) ImagePull(image interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ImagePull", reflect.TypeOf((*MockCommand)(nil).ImagePull), image)
}

// ImageTag mocks base method
func (m *MockCommand) ImageTag(imageID, repoTags string) error {
	ret := m.ctrl.Call(m, "ImageTag", imageID, repoTags)
	ret0, _ := ret[0].(error)
	return ret0
}

// ImageTag indicates an expected call of ImageTag
func (mr *MockCommandMockRecorder) ImageTag(imageID, repoTags interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ImageTag", reflect.TypeOf((*MockCommand)(nil).ImageTag), imageID, repoTags)
}

// Events mocks base method
func (m *MockCommand) Events(id, path string, evt chan dockercontroller.Event, services ...string) error {
	varargs := []interface{}{id, path, evt}
	for _, a := range services {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Events", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Events indicates an expected call of Events
func (mr *MockCommandMockRecorder) Events(id, path, evt interface{}, services ...interface{}) *gomock.Call {
	varargs := append([]interface{}{id, path, evt}, services...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Events", reflect.TypeOf((*MockCommand)(nil).Events), varargs...)
}

// UpWithEvent mocks base method
func (m *MockCommand) UpWithEvent(id, path, eventID string, evt chan dockercontroller.Event, services ...string) error {
	varargs := []interface{}{id, path, eventID, evt}
	for _, a := range services {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpWithEvent", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpWithEvent indicates an expected call of UpWithEvent
func (mr *MockCommandMockRecorder) UpWithEvent(id, path, eventID, evt interface{}, services ...interface{}) *gomock.Call {
	varargs := append([]interface{}{id, path, eventID, evt}, services...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpWithEvent", reflect.TypeOf((*MockCommand)(nil).UpWithEvent), varargs...)
}

// Info mocks base method
func (m *MockCommand) Info() (map[string]interface{}, error) {
	ret := m.ctrl.Call(m, "Info")
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Info indicates an expected call of Info
func (mr *MockCommandMockRecorder) Info() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockCommand)(nil).Info))
}

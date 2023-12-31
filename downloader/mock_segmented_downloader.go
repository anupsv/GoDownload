// Code generated by MockGen. DO NOT EDIT.
// Source: downloader/segmented_downloader.go

// Package downloader is a generated GoMock package.
package downloader

import (
	clients "GoDownload/clients"
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockSegmentManager is a mock of SegmentManager interface.
type MockSegmentManager struct {
	ctrl     *gomock.Controller
	recorder *MockSegmentManagerMockRecorder
}

// MockSegmentManagerMockRecorder is the mock recorder for MockSegmentManager.
type MockSegmentManagerMockRecorder struct {
	mock *MockSegmentManager
}

// NewMockSegmentManager creates a new mock instance.
func NewMockSegmentManager(ctrl *gomock.Controller) *MockSegmentManager {
	mock := &MockSegmentManager{ctrl: ctrl}
	mock.recorder = &MockSegmentManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSegmentManager) EXPECT() *MockSegmentManagerMockRecorder {
	return m.recorder
}

// DownloadSegment mocks base method.
func (m *MockSegmentManager) DownloadSegment(ctx context.Context, client clients.HttpClient, url string, start, end int64, destPath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadSegment", ctx, client, url, start, end, destPath)
	ret0, _ := ret[0].(error)
	return ret0
}

// DownloadSegment indicates an expected call of DownloadSegment.
func (mr *MockSegmentManagerMockRecorder) DownloadSegment(ctx, client, url, start, end, destPath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadSegment", reflect.TypeOf((*MockSegmentManager)(nil).DownloadSegment), ctx, client, url, start, end, destPath)
}

// MergeSegments mocks base method.
func (m *MockSegmentManager) MergeSegments(destPath string, segmentCount int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MergeSegments", destPath, segmentCount)
	ret0, _ := ret[0].(error)
	return ret0
}

// MergeSegments indicates an expected call of MergeSegments.
func (mr *MockSegmentManagerMockRecorder) MergeSegments(destPath, segmentCount interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MergeSegments", reflect.TypeOf((*MockSegmentManager)(nil).MergeSegments), destPath, segmentCount)
}

// MockSegmentManagerFactory is a mock of SegmentManagerFactory interface.
type MockSegmentManagerFactory struct {
	ctrl     *gomock.Controller
	recorder *MockSegmentManagerFactoryMockRecorder
}

// MockSegmentManagerFactoryMockRecorder is the mock recorder for MockSegmentManagerFactory.
type MockSegmentManagerFactoryMockRecorder struct {
	mock *MockSegmentManagerFactory
}

// NewMockSegmentManagerFactory creates a new mock instance.
func NewMockSegmentManagerFactory(ctrl *gomock.Controller) *MockSegmentManagerFactory {
	mock := &MockSegmentManagerFactory{ctrl: ctrl}
	mock.recorder = &MockSegmentManagerFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSegmentManagerFactory) EXPECT() *MockSegmentManagerFactoryMockRecorder {
	return m.recorder
}

// NewSegmentManager mocks base method.
func (m *MockSegmentManagerFactory) NewSegmentManager(client clients.HttpClient, url, destPath string) SegmentManager {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewSegmentManager", client, url, destPath)
	ret0, _ := ret[0].(SegmentManager)
	return ret0
}

// NewSegmentManager indicates an expected call of NewSegmentManager.
func (mr *MockSegmentManagerFactoryMockRecorder) NewSegmentManager(client, url, destPath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewSegmentManager", reflect.TypeOf((*MockSegmentManagerFactory)(nil).NewSegmentManager), client, url, destPath)
}

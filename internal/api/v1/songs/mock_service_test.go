// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Karaoke-Manager/karman/internal/service/song (interfaces: Service)

// Package songs is a generated GoMock package.
package songs

import (
	context "context"
	reflect "reflect"

	ultrastar "github.com/Karaoke-Manager/go-ultrastar"
	model "github.com/Karaoke-Manager/karman/internal/model"
	gomock "go.uber.org/mock/gomock"
)

// MockSongService is a mock of Service interface.
type MockSongService struct {
	ctrl     *gomock.Controller
	recorder *MockSongServiceMockRecorder
}

// MockSongServiceMockRecorder is the mock recorder for MockSongService.
type MockSongServiceMockRecorder struct {
	mock *MockSongService
}

// NewMockSongService creates a new mock instance.
func NewMockSongService(ctrl *gomock.Controller) *MockSongService {
	mock := &MockSongService{ctrl: ctrl}
	mock.recorder = &MockSongServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSongService) EXPECT() *MockSongServiceMockRecorder {
	return m.recorder
}

// CreateSong mocks base method.
func (m *MockSongService) CreateSong(arg0 context.Context, arg1 *ultrastar.Song) (model.Song, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSong", arg0, arg1)
	ret0, _ := ret[0].(model.Song)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSong indicates an expected call of CreateSong.
func (mr *MockSongServiceMockRecorder) CreateSong(arg0, arg1 interface{}) *ServiceCreateSongCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSong", reflect.TypeOf((*MockSongService)(nil).CreateSong), arg0, arg1)
	return &ServiceCreateSongCall{Call: call}
}

// ServiceCreateSongCall wrap *gomock.Call
type ServiceCreateSongCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ServiceCreateSongCall) Return(arg0 model.Song, arg1 error) *ServiceCreateSongCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ServiceCreateSongCall) Do(f func(context.Context, *ultrastar.Song) (model.Song, error)) *ServiceCreateSongCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ServiceCreateSongCall) DoAndReturn(f func(context.Context, *ultrastar.Song) (model.Song, error)) *ServiceCreateSongCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// DeleteSongByUUID mocks base method.
func (m *MockSongService) DeleteSongByUUID(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSongByUUID", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSongByUUID indicates an expected call of DeleteSongByUUID.
func (mr *MockSongServiceMockRecorder) DeleteSongByUUID(arg0, arg1 interface{}) *ServiceDeleteSongByUUIDCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSongByUUID", reflect.TypeOf((*MockSongService)(nil).DeleteSongByUUID), arg0, arg1)
	return &ServiceDeleteSongByUUIDCall{Call: call}
}

// ServiceDeleteSongByUUIDCall wrap *gomock.Call
type ServiceDeleteSongByUUIDCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ServiceDeleteSongByUUIDCall) Return(arg0 error) *ServiceDeleteSongByUUIDCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ServiceDeleteSongByUUIDCall) Do(f func(context.Context, string) error) *ServiceDeleteSongByUUIDCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ServiceDeleteSongByUUIDCall) DoAndReturn(f func(context.Context, string) error) *ServiceDeleteSongByUUIDCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FindSongs mocks base method.
func (m *MockSongService) FindSongs(arg0 context.Context, arg1, arg2 int) ([]model.Song, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindSongs", arg0, arg1, arg2)
	ret0, _ := ret[0].([]model.Song)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// FindSongs indicates an expected call of FindSongs.
func (mr *MockSongServiceMockRecorder) FindSongs(arg0, arg1, arg2 interface{}) *ServiceFindSongsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindSongs", reflect.TypeOf((*MockSongService)(nil).FindSongs), arg0, arg1, arg2)
	return &ServiceFindSongsCall{Call: call}
}

// ServiceFindSongsCall wrap *gomock.Call
type ServiceFindSongsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ServiceFindSongsCall) Return(arg0 []model.Song, arg1 int64, arg2 error) *ServiceFindSongsCall {
	c.Call = c.Call.Return(arg0, arg1, arg2)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ServiceFindSongsCall) Do(f func(context.Context, int, int) ([]model.Song, int64, error)) *ServiceFindSongsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ServiceFindSongsCall) DoAndReturn(f func(context.Context, int, int) ([]model.Song, int64, error)) *ServiceFindSongsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetSong mocks base method.
func (m *MockSongService) GetSong(arg0 context.Context, arg1 string) (model.Song, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSong", arg0, arg1)
	ret0, _ := ret[0].(model.Song)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSong indicates an expected call of GetSong.
func (mr *MockSongServiceMockRecorder) GetSong(arg0, arg1 interface{}) *ServiceGetSongCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSong", reflect.TypeOf((*MockSongService)(nil).GetSong), arg0, arg1)
	return &ServiceGetSongCall{Call: call}
}

// ServiceGetSongCall wrap *gomock.Call
type ServiceGetSongCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ServiceGetSongCall) Return(arg0 model.Song, arg1 error) *ServiceGetSongCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ServiceGetSongCall) Do(f func(context.Context, string) (model.Song, error)) *ServiceGetSongCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ServiceGetSongCall) DoAndReturn(f func(context.Context, string) (model.Song, error)) *ServiceGetSongCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetSongWithFiles mocks base method.
func (m *MockSongService) GetSongWithFiles(arg0 context.Context, arg1 string) (model.Song, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSongWithFiles", arg0, arg1)
	ret0, _ := ret[0].(model.Song)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSongWithFiles indicates an expected call of GetSongWithFiles.
func (mr *MockSongServiceMockRecorder) GetSongWithFiles(arg0, arg1 interface{}) *ServiceGetSongWithFilesCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSongWithFiles", reflect.TypeOf((*MockSongService)(nil).GetSongWithFiles), arg0, arg1)
	return &ServiceGetSongWithFilesCall{Call: call}
}

// ServiceGetSongWithFilesCall wrap *gomock.Call
type ServiceGetSongWithFilesCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ServiceGetSongWithFilesCall) Return(arg0 model.Song, arg1 error) *ServiceGetSongWithFilesCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ServiceGetSongWithFilesCall) Do(f func(context.Context, string) (model.Song, error)) *ServiceGetSongWithFilesCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ServiceGetSongWithFilesCall) DoAndReturn(f func(context.Context, string) (model.Song, error)) *ServiceGetSongWithFilesCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// SaveSong mocks base method.
func (m *MockSongService) SaveSong(arg0 context.Context, arg1 *model.Song) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveSong", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveSong indicates an expected call of SaveSong.
func (mr *MockSongServiceMockRecorder) SaveSong(arg0, arg1 interface{}) *ServiceSaveSongCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveSong", reflect.TypeOf((*MockSongService)(nil).SaveSong), arg0, arg1)
	return &ServiceSaveSongCall{Call: call}
}

// ServiceSaveSongCall wrap *gomock.Call
type ServiceSaveSongCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ServiceSaveSongCall) Return(arg0 error) *ServiceSaveSongCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ServiceSaveSongCall) Do(f func(context.Context, *model.Song) error) *ServiceSaveSongCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ServiceSaveSongCall) DoAndReturn(f func(context.Context, *model.Song) error) *ServiceSaveSongCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// UltraStarSong mocks base method.
func (m *MockSongService) UltraStarSong(arg0 context.Context, arg1 model.Song) *ultrastar.Song {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UltraStarSong", arg0, arg1)
	ret0, _ := ret[0].(*ultrastar.Song)
	return ret0
}

// UltraStarSong indicates an expected call of UltraStarSong.
func (mr *MockSongServiceMockRecorder) UltraStarSong(arg0, arg1 interface{}) *ServiceUltraStarSongCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UltraStarSong", reflect.TypeOf((*MockSongService)(nil).UltraStarSong), arg0, arg1)
	return &ServiceUltraStarSongCall{Call: call}
}

// ServiceUltraStarSongCall wrap *gomock.Call
type ServiceUltraStarSongCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ServiceUltraStarSongCall) Return(arg0 *ultrastar.Song) *ServiceUltraStarSongCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ServiceUltraStarSongCall) Do(f func(context.Context, model.Song) *ultrastar.Song) *ServiceUltraStarSongCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ServiceUltraStarSongCall) DoAndReturn(f func(context.Context, model.Song) *ultrastar.Song) *ServiceUltraStarSongCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

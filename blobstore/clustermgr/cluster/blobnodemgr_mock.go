// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cubefs/cubefs/blobstore/clustermgr/cluster (interfaces: BlobNodeManagerAPI)

// Package cluster is a generated GoMock package.
package cluster

import (
	context "context"
	reflect "reflect"

	clustermgr "github.com/cubefs/cubefs/blobstore/api/clustermgr"
	proto "github.com/cubefs/cubefs/blobstore/common/proto"
	gomock "github.com/golang/mock/gomock"
)

// MockBlobNodeManagerAPI is a mock of BlobNodeManagerAPI interface.
type MockBlobNodeManagerAPI struct {
	ctrl     *gomock.Controller
	recorder *MockBlobNodeManagerAPIMockRecorder
}

// MockBlobNodeManagerAPIMockRecorder is the mock recorder for MockBlobNodeManagerAPI.
type MockBlobNodeManagerAPIMockRecorder struct {
	mock *MockBlobNodeManagerAPI
}

// NewMockBlobNodeManagerAPI creates a new mock instance.
func NewMockBlobNodeManagerAPI(ctrl *gomock.Controller) *MockBlobNodeManagerAPI {
	mock := &MockBlobNodeManagerAPI{ctrl: ctrl}
	mock.recorder = &MockBlobNodeManagerAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBlobNodeManagerAPI) EXPECT() *MockBlobNodeManagerAPIMockRecorder {
	return m.recorder
}

// AddDisk mocks base method.
func (m *MockBlobNodeManagerAPI) AddDisk(arg0 context.Context, arg1 *clustermgr.BlobNodeDiskInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddDisk", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddDisk indicates an expected call of AddDisk.
func (mr *MockBlobNodeManagerAPIMockRecorder) AddDisk(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddDisk", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).AddDisk), arg0, arg1)
}

// AllocChunks mocks base method.
func (m *MockBlobNodeManagerAPI) AllocChunks(arg0 context.Context, arg1 AllocPolicy) ([]proto.DiskID, []proto.Vuid, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllocChunks", arg0, arg1)
	ret0, _ := ret[0].([]proto.DiskID)
	ret1, _ := ret[1].([]proto.Vuid)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// AllocChunks indicates an expected call of AllocChunks.
func (mr *MockBlobNodeManagerAPIMockRecorder) AllocChunks(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllocChunks", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).AllocChunks), arg0, arg1)
}

// AllocDiskID mocks base method.
func (m *MockBlobNodeManagerAPI) AllocDiskID(arg0 context.Context) (proto.DiskID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllocDiskID", arg0)
	ret0, _ := ret[0].(proto.DiskID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllocDiskID indicates an expected call of AllocDiskID.
func (mr *MockBlobNodeManagerAPIMockRecorder) AllocDiskID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllocDiskID", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).AllocDiskID), arg0)
}

// AllocNodeID mocks base method.
func (m *MockBlobNodeManagerAPI) AllocNodeID(arg0 context.Context) (proto.NodeID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllocNodeID", arg0)
	ret0, _ := ret[0].(proto.NodeID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllocNodeID indicates an expected call of AllocNodeID.
func (mr *MockBlobNodeManagerAPIMockRecorder) AllocNodeID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllocNodeID", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).AllocNodeID), arg0)
}

// CheckDiskInfoDuplicated mocks base method.
func (m *MockBlobNodeManagerAPI) CheckDiskInfoDuplicated(arg0 context.Context, arg1 proto.DiskID, arg2 *clustermgr.DiskInfo, arg3 *clustermgr.NodeInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckDiskInfoDuplicated", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckDiskInfoDuplicated indicates an expected call of CheckDiskInfoDuplicated.
func (mr *MockBlobNodeManagerAPIMockRecorder) CheckDiskInfoDuplicated(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckDiskInfoDuplicated", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).CheckDiskInfoDuplicated), arg0, arg1, arg2, arg3)
}

// CheckNodeInfoDuplicated mocks base method.
func (m *MockBlobNodeManagerAPI) CheckNodeInfoDuplicated(arg0 context.Context, arg1 *clustermgr.NodeInfo) (proto.NodeID, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckNodeInfoDuplicated", arg0, arg1)
	ret0, _ := ret[0].(proto.NodeID)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// CheckNodeInfoDuplicated indicates an expected call of CheckNodeInfoDuplicated.
func (mr *MockBlobNodeManagerAPIMockRecorder) CheckNodeInfoDuplicated(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckNodeInfoDuplicated", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).CheckNodeInfoDuplicated), arg0, arg1)
}

// GetDiskInfo mocks base method.
func (m *MockBlobNodeManagerAPI) GetDiskInfo(arg0 context.Context, arg1 proto.DiskID) (*clustermgr.BlobNodeDiskInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDiskInfo", arg0, arg1)
	ret0, _ := ret[0].(*clustermgr.BlobNodeDiskInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDiskInfo indicates an expected call of GetDiskInfo.
func (mr *MockBlobNodeManagerAPIMockRecorder) GetDiskInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDiskInfo", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).GetDiskInfo), arg0, arg1)
}

// GetHeartbeatChangeDisks mocks base method.
func (m *MockBlobNodeManagerAPI) GetHeartbeatChangeDisks() []HeartbeatEvent {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHeartbeatChangeDisks")
	ret0, _ := ret[0].([]HeartbeatEvent)
	return ret0
}

// GetHeartbeatChangeDisks indicates an expected call of GetHeartbeatChangeDisks.
func (mr *MockBlobNodeManagerAPIMockRecorder) GetHeartbeatChangeDisks() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeartbeatChangeDisks", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).GetHeartbeatChangeDisks))
}

// GetNodeInfo mocks base method.
func (m *MockBlobNodeManagerAPI) GetNodeInfo(arg0 context.Context, arg1 proto.NodeID) (*clustermgr.BlobNodeInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodeInfo", arg0, arg1)
	ret0, _ := ret[0].(*clustermgr.BlobNodeInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNodeInfo indicates an expected call of GetNodeInfo.
func (mr *MockBlobNodeManagerAPIMockRecorder) GetNodeInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodeInfo", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).GetNodeInfo), arg0, arg1)
}

// IsDiskWritable mocks base method.
func (m *MockBlobNodeManagerAPI) IsDiskWritable(arg0 context.Context, arg1 proto.DiskID) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDiskWritable", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsDiskWritable indicates an expected call of IsDiskWritable.
func (mr *MockBlobNodeManagerAPIMockRecorder) IsDiskWritable(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDiskWritable", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).IsDiskWritable), arg0, arg1)
}

// IsDroppingDisk mocks base method.
func (m *MockBlobNodeManagerAPI) IsDroppingDisk(arg0 context.Context, arg1 proto.DiskID) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDroppingDisk", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsDroppingDisk indicates an expected call of IsDroppingDisk.
func (mr *MockBlobNodeManagerAPIMockRecorder) IsDroppingDisk(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDroppingDisk", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).IsDroppingDisk), arg0, arg1)
}

// ListDiskInfo mocks base method.
func (m *MockBlobNodeManagerAPI) ListDiskInfo(arg0 context.Context, arg1 *clustermgr.ListOptionArgs) ([]*clustermgr.BlobNodeDiskInfo, proto.DiskID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListDiskInfo", arg0, arg1)
	ret0, _ := ret[0].([]*clustermgr.BlobNodeDiskInfo)
	ret1, _ := ret[1].(proto.DiskID)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ListDiskInfo indicates an expected call of ListDiskInfo.
func (mr *MockBlobNodeManagerAPIMockRecorder) ListDiskInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDiskInfo", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).ListDiskInfo), arg0, arg1)
}

// ListDroppingDisk mocks base method.
func (m *MockBlobNodeManagerAPI) ListDroppingDisk(arg0 context.Context) ([]*clustermgr.BlobNodeDiskInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListDroppingDisk", arg0)
	ret0, _ := ret[0].([]*clustermgr.BlobNodeDiskInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListDroppingDisk indicates an expected call of ListDroppingDisk.
func (mr *MockBlobNodeManagerAPIMockRecorder) ListDroppingDisk(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDroppingDisk", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).ListDroppingDisk), arg0)
}

// RefreshExpireTime mocks base method.
func (m *MockBlobNodeManagerAPI) RefreshExpireTime() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RefreshExpireTime")
}

// RefreshExpireTime indicates an expected call of RefreshExpireTime.
func (mr *MockBlobNodeManagerAPIMockRecorder) RefreshExpireTime() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshExpireTime", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).RefreshExpireTime))
}

// SetStatus mocks base method.
func (m *MockBlobNodeManagerAPI) SetStatus(arg0 context.Context, arg1 proto.DiskID, arg2 proto.DiskStatus, arg3 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetStatus", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetStatus indicates an expected call of SetStatus.
func (mr *MockBlobNodeManagerAPIMockRecorder) SetStatus(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetStatus", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).SetStatus), arg0, arg1, arg2, arg3)
}

// Stat mocks base method.
func (m *MockBlobNodeManagerAPI) Stat(arg0 context.Context, arg1 proto.DiskType) *clustermgr.SpaceStatInfo {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stat", arg0, arg1)
	ret0, _ := ret[0].(*clustermgr.SpaceStatInfo)
	return ret0
}

// Stat indicates an expected call of Stat.
func (mr *MockBlobNodeManagerAPIMockRecorder) Stat(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stat", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).Stat), arg0, arg1)
}

// ValidateNodeInfo mocks base method.
func (m *MockBlobNodeManagerAPI) ValidateNodeInfo(arg0 context.Context, arg1 *clustermgr.NodeInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateNodeInfo", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateNodeInfo indicates an expected call of ValidateNodeInfo.
func (mr *MockBlobNodeManagerAPIMockRecorder) ValidateNodeInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateNodeInfo", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).ValidateNodeInfo), arg0, arg1)
}

// addDiskNoLocked mocks base method.
func (m *MockBlobNodeManagerAPI) addDiskNoLocked(arg0 *diskItem) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "addDiskNoLocked", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// addDiskNoLocked indicates an expected call of addDiskNoLocked.
func (mr *MockBlobNodeManagerAPIMockRecorder) addDiskNoLocked(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "addDiskNoLocked", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).addDiskNoLocked), arg0)
}

// addDroppingDisk mocks base method.
func (m *MockBlobNodeManagerAPI) addDroppingDisk(arg0 proto.DiskID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "addDroppingDisk", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// addDroppingDisk indicates an expected call of addDroppingDisk.
func (mr *MockBlobNodeManagerAPIMockRecorder) addDroppingDisk(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "addDroppingDisk", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).addDroppingDisk), arg0)
}

// addDroppingNode mocks base method.
func (m *MockBlobNodeManagerAPI) addDroppingNode(arg0 proto.NodeID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "addDroppingNode", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// addDroppingNode indicates an expected call of addDroppingNode.
func (mr *MockBlobNodeManagerAPIMockRecorder) addDroppingNode(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "addDroppingNode", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).addDroppingNode), arg0)
}

// droppedDisk mocks base method.
func (m *MockBlobNodeManagerAPI) droppedDisk(arg0 proto.DiskID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "droppedDisk", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// droppedDisk indicates an expected call of droppedDisk.
func (mr *MockBlobNodeManagerAPIMockRecorder) droppedDisk(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "droppedDisk", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).droppedDisk), arg0)
}

// droppedNode mocks base method.
func (m *MockBlobNodeManagerAPI) droppedNode(arg0 proto.NodeID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "droppedNode", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// droppedNode indicates an expected call of droppedNode.
func (mr *MockBlobNodeManagerAPIMockRecorder) droppedNode(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "droppedNode", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).droppedNode), arg0)
}

// isDroppingDisk mocks base method.
func (m *MockBlobNodeManagerAPI) isDroppingDisk(arg0 proto.DiskID) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "isDroppingDisk", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// isDroppingDisk indicates an expected call of isDroppingDisk.
func (mr *MockBlobNodeManagerAPIMockRecorder) isDroppingDisk(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "isDroppingDisk", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).isDroppingDisk), arg0)
}

// isDroppingNode mocks base method.
func (m *MockBlobNodeManagerAPI) isDroppingNode(arg0 proto.NodeID) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "isDroppingNode", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// isDroppingNode indicates an expected call of isDroppingNode.
func (mr *MockBlobNodeManagerAPIMockRecorder) isDroppingNode(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "isDroppingNode", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).isDroppingNode), arg0)
}

// updateDiskNoLocked mocks base method.
func (m *MockBlobNodeManagerAPI) updateDiskNoLocked(arg0 *diskItem) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "updateDiskNoLocked", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// updateDiskNoLocked indicates an expected call of updateDiskNoLocked.
func (mr *MockBlobNodeManagerAPIMockRecorder) updateDiskNoLocked(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "updateDiskNoLocked", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).updateDiskNoLocked), arg0)
}

// updateDiskStatusNoLocked mocks base method.
func (m *MockBlobNodeManagerAPI) updateDiskStatusNoLocked(arg0 proto.DiskID, arg1 proto.DiskStatus) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "updateDiskStatusNoLocked", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// updateDiskStatusNoLocked indicates an expected call of updateDiskStatusNoLocked.
func (mr *MockBlobNodeManagerAPIMockRecorder) updateDiskStatusNoLocked(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "updateDiskStatusNoLocked", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).updateDiskStatusNoLocked), arg0, arg1)
}

// updateNodeNoLocked mocks base method.
func (m *MockBlobNodeManagerAPI) updateNodeNoLocked(arg0 *nodeItem) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "updateNodeNoLocked", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// updateNodeNoLocked indicates an expected call of updateNodeNoLocked.
func (mr *MockBlobNodeManagerAPIMockRecorder) updateNodeNoLocked(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "updateNodeNoLocked", reflect.TypeOf((*MockBlobNodeManagerAPI)(nil).updateNodeNoLocked), arg0)
}

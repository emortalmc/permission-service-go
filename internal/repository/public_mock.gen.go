// Code generated by MockGen. DO NOT EDIT.
// Source: internal/repository/public.go

// Package repository is a generated GoMock package.
package repository

import (
	context "context"
	model "permission-service-go/internal/repository/model"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// AddRoleToPlayer mocks base method.
func (m *MockRepository) AddRoleToPlayer(ctx context.Context, playerId uuid.UUID, roleId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddRoleToPlayer", ctx, playerId, roleId)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddRoleToPlayer indicates an expected call of AddRoleToPlayer.
func (mr *MockRepositoryMockRecorder) AddRoleToPlayer(ctx, playerId, roleId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddRoleToPlayer", reflect.TypeOf((*MockRepository)(nil).AddRoleToPlayer), ctx, playerId, roleId)
}

// CreateRole mocks base method.
func (m *MockRepository) CreateRole(ctx context.Context, role *model.Role) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRole", ctx, role)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateRole indicates an expected call of CreateRole.
func (mr *MockRepositoryMockRecorder) CreateRole(ctx, role interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRole", reflect.TypeOf((*MockRepository)(nil).CreateRole), ctx, role)
}

// DoesRoleExist mocks base method.
func (m *MockRepository) DoesRoleExist(ctx context.Context, roleId string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DoesRoleExist", ctx, roleId)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DoesRoleExist indicates an expected call of DoesRoleExist.
func (mr *MockRepositoryMockRecorder) DoesRoleExist(ctx, roleId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DoesRoleExist", reflect.TypeOf((*MockRepository)(nil).DoesRoleExist), ctx, roleId)
}

// GetPlayerRoleIds mocks base method.
func (m *MockRepository) GetPlayerRoleIds(ctx context.Context, playerId uuid.UUID) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPlayerRoleIds", ctx, playerId)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPlayerRoleIds indicates an expected call of GetPlayerRoleIds.
func (mr *MockRepositoryMockRecorder) GetPlayerRoleIds(ctx, playerId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPlayerRoleIds", reflect.TypeOf((*MockRepository)(nil).GetPlayerRoleIds), ctx, playerId)
}

// GetRole mocks base method.
func (m *MockRepository) GetRole(ctx context.Context, roleId string) (*model.Role, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRole", ctx, roleId)
	ret0, _ := ret[0].(*model.Role)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRole indicates an expected call of GetRole.
func (mr *MockRepositoryMockRecorder) GetRole(ctx, roleId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRole", reflect.TypeOf((*MockRepository)(nil).GetRole), ctx, roleId)
}

// GetRoles mocks base method.
func (m *MockRepository) GetRoles(ctx context.Context) ([]*model.Role, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRoles", ctx)
	ret0, _ := ret[0].([]*model.Role)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRoles indicates an expected call of GetRoles.
func (mr *MockRepositoryMockRecorder) GetRoles(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRoles", reflect.TypeOf((*MockRepository)(nil).GetRoles), ctx)
}

// RemoveRoleFromPlayer mocks base method.
func (m *MockRepository) RemoveRoleFromPlayer(ctx context.Context, playerId uuid.UUID, roleId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveRoleFromPlayer", ctx, playerId, roleId)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveRoleFromPlayer indicates an expected call of RemoveRoleFromPlayer.
func (mr *MockRepositoryMockRecorder) RemoveRoleFromPlayer(ctx, playerId, roleId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveRoleFromPlayer", reflect.TypeOf((*MockRepository)(nil).RemoveRoleFromPlayer), ctx, playerId, roleId)
}

// UpdateRole mocks base method.
func (m *MockRepository) UpdateRole(ctx context.Context, newRole *model.Role) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateRole", ctx, newRole)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateRole indicates an expected call of UpdateRole.
func (mr *MockRepositoryMockRecorder) UpdateRole(ctx, newRole interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateRole", reflect.TypeOf((*MockRepository)(nil).UpdateRole), ctx, newRole)
}

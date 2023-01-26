package service

import (
	"context"
	permService "github.com/emortalmc/proto-specs/gen/go/grpc/permission"
	protoModel "github.com/emortalmc/proto-specs/gen/go/model/permission"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"permission-service-go/internal/repository"
	"permission-service-go/internal/repository/model"
	"testing"
)

func TestPermissionService_GetAllRoles(t *testing.T) {
	mockCntrl := gomock.NewController(t)
	mockRepo := repository.NewMockRepository(mockCntrl)
	mockRoles := genericRoles()
	mockProtoRoles := genericProtoRoles()

	mockRepo.EXPECT().GetRoles(context.Background()).Return(mockRoles, nil)

	svc := permissionService{
		repo: mockRepo,
	}

	response, err := svc.GetAllRoles(context.Background(), &permService.GetAllRolesRequest{})
	assert.NoError(t, err)
	assert.Equal(t, response, &permService.GetAllRolesResponse{
		Roles: mockProtoRoles,
	})
}

func stringPointer(s string) *string {
	return &s
}

func genericRoles() []*model.Role {
	return []*model.Role{
		{
			Id:            "1",
			Priority:      1,
			DisplayPrefix: stringPointer("testPrefix"),
			DisplayName:   stringPointer("testName"),
			Permissions: []model.PermissionNode{
				{
					Node:            "testNode",
					PermissionState: protoModel.PermissionNode_ALLOW,
				},
			},
		},
		{
			Id:            "2",
			Priority:      100,
			DisplayPrefix: nil,
			DisplayName:   nil,
			Permissions: []model.PermissionNode{
				{
					Node:            "testNode",
					PermissionState: protoModel.PermissionNode_DENY,
				},
			},
		},
	}
}

func genericProtoRoles() []*protoModel.Role {
	roles := genericRoles()
	protoRoles := make([]*protoModel.Role, len(roles))
	for i, role := range genericRoles() {
		protoRoles[i] = role.ToProto()
	}
	return protoRoles
}

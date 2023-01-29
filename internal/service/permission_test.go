package service

import (
	"context"
	permService "github.com/emortalmc/proto-specs/gen/go/grpc/permission"
	"github.com/emortalmc/proto-specs/gen/go/message/permission"
	protoModel "github.com/emortalmc/proto-specs/gen/go/model/permission"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"permission-service-go/internal/notifier"
	"permission-service-go/internal/repository"
	"permission-service-go/internal/repository/model"
	"testing"
)

// TODO: Have the service sanity check incoming requests for missing fields

func TestPermissionService_GetAllRoles(t *testing.T) {
	mockCntrl := gomock.NewController(t)
	mockRepo := repository.NewMockRepository(mockCntrl)
	mockRoles := []*model.Role{createGenericRole(), createPartialGenericRole()}
	mockProtoRoles := []*protoModel.Role{createGenericRole().ToProto(), createPartialGenericRole().ToProto()}

	mockRepo.EXPECT().GetRoles(context.Background()).Return(mockRoles, nil)

	svc := permissionService{
		repo: mockRepo,
	}

	expected := &permService.GetAllRolesResponse{Roles: mockProtoRoles}

	response, err := svc.GetAllRoles(context.Background(), &permService.GetAllRolesRequest{})
	assert.NoError(t, err)
	assert.Equal(t, expected, response)

	// Test with no roles returned
	mockRepo.EXPECT().GetRoles(context.Background()).Return(nil, nil)
	response, err = svc.GetAllRoles(context.Background(), &permService.GetAllRolesRequest{})
	assert.NoError(t, err)
	assert.Equal(t, &permService.GetAllRolesResponse{}, response)
}

var testUserIds = []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

func TestPermissionService_GetPlayerRoles(t *testing.T) {
	mockCntrl := gomock.NewController(t)
	mockRepo := repository.NewMockRepository(mockCntrl)

	testRoleIds := []string{"default", "testRole"}

	mockRepo.EXPECT().GetPlayerRoleIds(context.Background(), testUserIds[0]).Return(testRoleIds, nil)

	svc := permissionService{
		repo: mockRepo,
	}

	response, err := svc.GetPlayerRoles(context.Background(), &permService.GetPlayerRolesRequest{
		PlayerId: testUserIds[0].String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, response, &permService.PlayerRolesResponse{RoleIds: testRoleIds})

	// Test with only default role
	defaultTestRoleIds := []string{"default"}

	mockRepo.EXPECT().GetPlayerRoleIds(context.Background(), testUserIds[1]).Return(defaultTestRoleIds, nil)
	response, err = svc.GetPlayerRoles(context.Background(), &permService.GetPlayerRolesRequest{
		PlayerId: testUserIds[1].String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, response, &permService.PlayerRolesResponse{RoleIds: defaultTestRoleIds})
}

func TestPermissionService_CreateRole(t *testing.T) {
	mockCntrl := gomock.NewController(t)
	mockRepo := repository.NewMockRepository(mockCntrl)
	mockNotifier := notifier.NewMockNotifier(mockCntrl)

	role := createGenericRole()
	role.Permissions = make([]model.PermissionNode, 0)

	// Test successful role creation
	mockRepo.EXPECT().CreateRole(context.Background(), role).Return(nil)
	mockNotifier.EXPECT().RoleUpdate(context.Background(), role, permission.RoleUpdateMessage_CREATE).Return(nil)

	svc := permissionService{
		repo:  mockRepo,
		notif: mockNotifier,
	}

	response, err := svc.CreateRole(context.Background(), &permService.RoleCreateRequest{
		Id:            role.Id,
		Priority:      role.Priority,
		DisplayName:   role.DisplayName,
		DisplayPrefix: role.DisplayPrefix,
	})

	assert.NoError(t, err)
	assert.Equal(t, response, &permService.CreateRoleResponse{
		Role: role.ToProto(),
	})
}

// Test with partial role
func TestPermissionService_CreateRole2(t *testing.T) {
	mockCntrl := gomock.NewController(t)
	mockRepo := repository.NewMockRepository(mockCntrl)
	mockNotifier := notifier.NewMockNotifier(mockCntrl)

	role := createPartialGenericRole()
	role.Permissions = make([]model.PermissionNode, 0)

	mockRepo.EXPECT().CreateRole(context.Background(), role).Return(nil)
	mockNotifier.EXPECT().RoleUpdate(context.Background(), role, permission.RoleUpdateMessage_CREATE).Return(nil)

	svc := permissionService{
		repo:  mockRepo,
		notif: mockNotifier,
	}

	response, err := svc.CreateRole(context.Background(), &permService.RoleCreateRequest{
		Id:            role.Id,
		Priority:      role.Priority,
		DisplayName:   role.DisplayName,
		DisplayPrefix: role.DisplayPrefix,
	})
	assert.NoError(t, err)
	assert.Equal(t, response, &permService.CreateRoleResponse{
		Role: role.ToProto(),
	})
}

// Test with existing role
func TestPermissionService_CreateRole3(t *testing.T) {
	mockCntrl := gomock.NewController(t)
	mockRepo := repository.NewMockRepository(mockCntrl)
	mockNotifier := notifier.NewMockNotifier(mockCntrl)

	mockRole := createGenericRole()
	mockRole.Permissions = make([]model.PermissionNode, 0)

	dupEx := mongo.WriteException{
		WriteErrors: []mongo.WriteError{
			{
				Index:   1,
				Code:    11000,
				Message: "duplicate key error",
			},
		},
	}
	assert.True(t, mongo.IsDuplicateKeyError(dupEx))
	mockRepo.EXPECT().CreateRole(context.Background(), mockRole).Return(dupEx)

	svc := permissionService{
		repo:  mockRepo,
		notif: mockNotifier,
	}

	response, err := svc.CreateRole(context.Background(), &permService.RoleCreateRequest{
		Id:            mockRole.Id,
		Priority:      mockRole.Priority,
		DisplayName:   mockRole.DisplayName,
		DisplayPrefix: mockRole.DisplayPrefix,
	})
	assert.Error(t, err)
	assert.True(t, status.Convert(err).Code() == codes.AlreadyExists)
	assert.Nil(t, response)
}

type updateRoleTest struct {
	// dbRole will be returned by the mock repository
	dbRole *model.Role

	getRoleErr    error
	updateRoleErr error

	mockReq *permService.RoleUpdateRequest

	// expectedUpdatedDbRole is the role that should be passed to the DB as UpdateRole by the service
	expectedUpdatedDbRole *model.Role
	notifChangeType       *permission.RoleUpdateMessage_ChangeType

	expectedErr func(t *testing.T, err error) bool
	expectedRes *permService.UpdateRoleResponse
}

// TODO YOU ARE HERE
var updateRoleTests = map[string]updateRoleTest{
	"successful update": {
		dbRole: createGenericRole(),

		getRoleErr:    nil,
		updateRoleErr: nil,

		mockReq: &permService.RoleUpdateRequest{
			Id:               createGenericRole().Id,
			Priority:         pointer(uint32(10000)),
			DisplayName:      pointer("new display name"),
			DisplayPrefix:    pointer("new display prefix"),
			UnsetPermissions: []string{""},
		},

		expectedUpdatedDbRole: &model.Role{
			Id:            createGenericRole().Id,
			Priority:      10000,
			DisplayName:   pointer("new display name"),
			DisplayPrefix: pointer("new display prefix"),
			Permissions:   createGenericRole().Permissions,
		},
		notifChangeType: pointer(permission.RoleUpdateMessage_MODIFY),

		expectedErr: nil,
		expectedRes: &permService.UpdateRoleResponse{
			Role: &protoModel.Role{
				Id:            createGenericRole().Id,
				Priority:      10000,
				DisplayName:   pointer("new display name"),
				DisplayPrefix: pointer("new display prefix"),
				Permissions:   createGenericRole().ToProto().Permissions,
			},
		},
	},
	"successful update with changed permissions": {
		dbRole: &model.Role{
			Id:            "test-role",
			Priority:      100,
			DisplayName:   pointer("test role"),
			DisplayPrefix: pointer("test"),
			Permissions: []model.PermissionNode{
				{
					Node:  "test.permission",
					State: protoModel.PermissionNode_ALLOW,
				},
				{
					Node:  "test.permission2",
					State: protoModel.PermissionNode_DENY,
				},
				{
					Node:  "test.permission3",
					State: protoModel.PermissionNode_DENY,
				},
			},
		},

		getRoleErr:    nil,
		updateRoleErr: nil,

		mockReq: &permService.RoleUpdateRequest{
			Id:               "test-role",
			Priority:         pointer(uint32(10000)),
			DisplayName:      pointer("<rainbow><username><\\rainbow>"),
			DisplayPrefix:    pointer("<red><bold>"),
			UnsetPermissions: []string{"test.permission2"},
			SetPermissions: []*protoModel.PermissionNode{
				{
					Node:  "test.permission3", // Tests overwriting existing permission
					State: protoModel.PermissionNode_ALLOW,
				},
				{
					Node:  "test.permission4",
					State: protoModel.PermissionNode_DENY,
				},
			},
		},

		expectedUpdatedDbRole: &model.Role{
			Id:            "test-role",
			Priority:      10000,
			DisplayName:   pointer("<rainbow><username><\\rainbow>"),
			DisplayPrefix: pointer("<red><bold>"),
			Permissions: []model.PermissionNode{
				{
					Node:  "test.permission",
					State: protoModel.PermissionNode_ALLOW,
				},
				{
					Node:  "test.permission3",
					State: protoModel.PermissionNode_ALLOW,
				},
				{
					Node:  "test.permission4",
					State: protoModel.PermissionNode_DENY,
				},
			},
		},
		notifChangeType: pointer(permission.RoleUpdateMessage_MODIFY),

		// TODO can we go over this. I think there's some verification with overwriting perms we're not doing
		expectedRes: &permService.UpdateRoleResponse{
			Role: &protoModel.Role{
				Id:            "test-role",
				Priority:      10000,
				DisplayName:   pointer("<rainbow><username><\\rainbow>"),
				DisplayPrefix: pointer("<red><bold>"),
				Permissions: []*protoModel.PermissionNode{
					{
						Node:  "test.permission",
						State: protoModel.PermissionNode_ALLOW,
					},
					{
						Node:  "test.permission3",
						State: protoModel.PermissionNode_ALLOW,
					},
					{
						Node:  "test.permission4",
						State: protoModel.PermissionNode_DENY,
					},
				},
			},
		},
		expectedErr: nil,
	},
	"role_doesnt_exist": {
		dbRole:     &model.Role{Id: "test-role"},
		getRoleErr: mongo.ErrNoDocuments,

		updateRoleErr:         nil,
		expectedUpdatedDbRole: nil,

		mockReq: &permService.RoleUpdateRequest{
			Id:               "test-role",
			UnsetPermissions: []string{""},
		},

		expectedErr: func(t *testing.T, err error) bool {
			return status.Code(err) == codes.NotFound
		},
		expectedRes: nil,
	},
}

// TODO: Note errors are probably because of notifier mocks right now
func TestPermissionService_UpdateRole(t *testing.T) {
	for name, test := range updateRoleTests {
		t.Run(name, func(t *testing.T) {
			mockCntrl := gomock.NewController(t)
			mockRepo := repository.NewMockRepository(mockCntrl)
			mockNotifier := notifier.NewMockNotifier(mockCntrl)

			svc := permissionService{
				repo:  mockRepo,
				notif: mockNotifier,
			}

			if test.dbRole != nil {
				mockRepo.EXPECT().GetRole(context.Background(), test.dbRole.Id).Return(test.dbRole, test.getRoleErr)
			}
			if test.expectedUpdatedDbRole != nil {
				mockRepo.EXPECT().UpdateRole(context.Background(), test.expectedUpdatedDbRole).Return(test.updateRoleErr)
			}
			if test.notifChangeType != nil {
				mockNotifier.EXPECT().RoleUpdate(context.Background(), test.expectedUpdatedDbRole, *test.notifChangeType).Return(nil)
			}

			response, err := svc.UpdateRole(context.Background(), test.mockReq)
			if test.expectedErr != nil {
				assert.True(t, test.expectedErr(t, err))
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedRes, response)
		})
	}
}

type addRoleToPlayerTest struct {
	roleExists bool

	addRoleErr error // e.g AlreadyHasRoleError

	expectedErr func(t *testing.T, err error)
}

var addRoleToPlayerTests = map[string]addRoleToPlayerTest{
	"valid": {
		roleExists: true,
		addRoleErr: nil,
		expectedErr: nil,
	},
	"role_doesnt_exist": {
		roleExists: false,
		addRoleErr: nil,
		expectedErr: func(t *testing.T, err error) {
			assert.Equal(t, codes.NotFound, status.Code(err))

			detailArr := status.Convert(err).Details()
			assert.Len(t, detailArr, 1)

			details := detailArr[0].(*permService.AddRoleToPlayerError)
			assert.Equal(t, permService.AddRoleToPlayerError_ROLE_NOT_FOUND, details.ErrorType)
		},
	},
	"already_has_role": {
		roleExists: true,
		addRoleErr: repository.AlreadyHasRoleError,
		expectedErr: func(t *testing.T, err error) {
			assert.Equal(t, codes.AlreadyExists, status.Code(err))

			detailArr := status.Convert(err).Details()
			assert.Len(t, detailArr, 1)

			details := detailArr[0].(*permService.AddRoleToPlayerError)
			assert.Equal(t, permService.AddRoleToPlayerError_ALREADY_HAS_ROLE, details.ErrorType)
		},
	},
}

// Valid add role to player request
func TestPermissionService_AddRoleToPlayer(t *testing.T) {
	for name, test := range addRoleToPlayerTests {
		t.Run(name, func(t *testing.T) {
			mockCntrl := gomock.NewController(t)
			mockRepo := repository.NewMockRepository(mockCntrl)
			mockNotifier := notifier.NewMockNotifier(mockCntrl)

			svc := permissionService{
				repo:  mockRepo,
				notif: mockNotifier,
			}

			playerId := uuid.New()
			playerIdStr := playerId.String()
			roleId := "test-role"

			mockRepo.EXPECT().DoesRoleExist(context.Background(), roleId).Return(test.roleExists, nil)
			if test.roleExists {
				mockRepo.EXPECT().AddRoleToPlayer(context.Background(), playerId, roleId).Return(test.addRoleErr)

				if test.addRoleErr == nil {
					mockNotifier.EXPECT().PlayerRolesUpdate(context.Background(), playerIdStr, roleId, permission.PlayerRolesUpdateMessage_ADD).Return(nil)
				}
			}

			_, err := svc.AddRoleToPlayer(context.Background(), &permService.AddRoleToPlayerRequest{
				RoleId:   roleId,
				PlayerId: playerIdStr,
			})

			if test.expectedErr != nil {
				test.expectedErr(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type removeRoleFromPlayerTest struct {
	removeRoleErr error // e.g DoesNotHaveRoleError, mongo.ErrNoDocuments

	expectedErr func(t *testing.T, err error)
}

var removeRoleFromPlayerTests = map[string]removeRoleFromPlayerTest{
	"valid": {
		removeRoleErr: nil,
		expectedErr:   nil,
	},
	"doesnt_have_role": {
		removeRoleErr: repository.DoesNotHaveRoleError,
		expectedErr: func(t *testing.T, err error) {
			assert.Equal(t, codes.NotFound, status.Code(err))

			detailArr := status.Convert(err).Details()
			assert.Len(t, detailArr, 1)

			details := detailArr[0].(*permService.RemoveRoleFromPlayerError)
			assert.Equal(t, permService.RemoveRoleFromPlayerError_DOES_NOT_HAVE_ROLE, details.ErrorType)
		},
	},
	"player_not_found": {
		removeRoleErr: mongo.ErrNoDocuments,
		expectedErr: func(t *testing.T, err error) {
			assert.Equal(t, codes.NotFound, status.Code(err))

			detailArr := status.Convert(err).Details()
			assert.Len(t, detailArr, 1)

			details := detailArr[0].(*permService.RemoveRoleFromPlayerError)
			assert.Equal(t, permService.RemoveRoleFromPlayerError_PLAYER_NOT_FOUND, details.ErrorType)
		},
	},
}

func TestPermissionService_RemoveRoleFromPlayer(t *testing.T) {
	for name, test := range removeRoleFromPlayerTests {
		t.Run(name, func(t *testing.T) {
			mockCntrl := gomock.NewController(t)
			mockRepo := repository.NewMockRepository(mockCntrl)
			mockNotifier := notifier.NewMockNotifier(mockCntrl)

			svc := permissionService{
				repo:  mockRepo,
				notif: mockNotifier,
			}

			playerId := uuid.New()
			playerIdStr := playerId.String()
			roleId := "test-role"

			mockRepo.EXPECT().RemoveRoleFromPlayer(context.Background(), playerId, roleId).Return(test.removeRoleErr)

			if test.removeRoleErr == nil {
				mockNotifier.EXPECT().PlayerRolesUpdate(context.Background(), playerIdStr, roleId, permission.PlayerRolesUpdateMessage_REMOVE).Return(nil)
			}

			_, err := svc.RemoveRoleFromPlayer(context.Background(), &permService.RemoveRoleFromPlayerRequest{
				RoleId:   roleId,
				PlayerId: playerIdStr,
			})

			if test.expectedErr != nil {
				test.expectedErr(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func pointer[T any](t T) *T {
	return &t
}

func createGenericRole() *model.Role {
	return &model.Role{
		Id:            "1",
		Priority:      1,
		DisplayPrefix: pointer("testPrefix"),
		DisplayName:   pointer("testName"),
		Permissions: []model.PermissionNode{
			{
				Node:  "testNode",
				State: protoModel.PermissionNode_ALLOW,
			},
		},
	}
}

func createPartialGenericRole() *model.Role {
	return &model.Role{
		Id:            "2",
		Priority:      100,
		DisplayPrefix: nil,
		DisplayName:   nil,
		Permissions: []model.PermissionNode{
			{
				Node:  "testNode",
				State: protoModel.PermissionNode_DENY,
			},
		},
	}
}

var mongoDupeErr = mongo.WriteException{
	WriteErrors: []mongo.WriteError{
		{
			Code: 11000,
		},
	},
}

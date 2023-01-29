package service

import (
	"context"
	"fmt"
	"github.com/emortalmc/proto-specs/gen/go/grpc/permission"
	permission2 "github.com/emortalmc/proto-specs/gen/go/message/permission"
	protoModel "github.com/emortalmc/proto-specs/gen/go/model/permission"
	"github.com/google/uuid"
	mongoDb "go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"permission-service-go/internal/notifier"
	"permission-service-go/internal/repository"
	"permission-service-go/internal/repository/model"
)

type permissionService struct {
	permission.PermissionServiceServer
	repo  repository.Repository
	notif notifier.Notifier
}

func NewPermissionService(repo repository.Repository, notif notifier.Notifier) permission.PermissionServiceServer {
	return &permissionService{
		repo:  repo,
		notif: notif,
	}
}

func (s *permissionService) GetAllRoles(ctx context.Context, _ *permission.GetAllRolesRequest) (*permission.GetAllRolesResponse, error) {
	roles, err := s.repo.GetRoles(ctx)
	if err != nil {
		return nil, err
	}

	if roles == nil {
		return &permission.GetAllRolesResponse{}, nil
	}

	var protoRoles = make([]*protoModel.Role, len(roles))
	for i, role := range roles {
		protoRoles[i] = role.ToProto()
	}

	return &permission.GetAllRolesResponse{
		Roles: protoRoles,
	}, nil
}

func (s *permissionService) GetPlayerRoles(ctx context.Context, req *permission.GetPlayerRolesRequest) (*permission.PlayerRolesResponse, error) {
	pId, err := uuid.Parse(req.PlayerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid player id %s", req.PlayerId))
	}
	roles, err := s.repo.GetPlayerRoleIds(ctx, pId)
	if err != nil {
		return nil, err
	}

	return &permission.PlayerRolesResponse{
		RoleIds: roles,
	}, nil
}

func (s *permissionService) CreateRole(ctx context.Context, req *permission.RoleCreateRequest) (*permission.CreateRoleResponse, error) {
	role := &model.Role{
		Id:            req.Id,
		Priority:      req.Priority,
		DisplayPrefix: req.DisplayPrefix,
		DisplayName:   req.DisplayName,
		Permissions:   make([]model.PermissionNode, 0),
	}

	err := s.repo.CreateRole(ctx, role)

	if err != nil {
		if mongoDb.IsDuplicateKeyError(err) {
			return nil, status.Error(codes.AlreadyExists, "role already exists")
		}
		zap.S().Errorw("error creating role", "error", err)
		return nil, err
	}

	err = s.notif.RoleUpdate(ctx, role, permission2.RoleUpdateMessage_CREATE)
	if err != nil {
		zap.S().Errorw("error sending role update notification", "error", err)
	}

	return &permission.CreateRoleResponse{
		Role: role.ToProto(),
	}, err
}

func (s *permissionService) UpdateRole(ctx context.Context, req *permission.RoleUpdateRequest) (*permission.UpdateRoleResponse, error) {
	role, err := s.repo.GetRole(ctx, req.Id)

	if err != nil {
		if err == mongoDb.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "Role not found")
		}
		zap.S().Errorw("error getting role", "error", err)
		return nil, err
	}

	if req.Priority != nil {
		role.Priority = *req.Priority
	}
	if req.DisplayPrefix != nil {
		role.DisplayPrefix = req.DisplayPrefix
	}
	if req.DisplayName != nil {
		role.DisplayName = req.DisplayName
	}

	for _, perm := range req.UnsetPermissions {
		for i, node := range role.Permissions {
			if node.Node == perm {
				role.Permissions = append(role.Permissions[:i], role.Permissions[i+1:]...)
			}
		}
	}

	// Update the permission state if it already exists, otherwise add it
	for _, perm := range req.SetPermissions {
		existed := false
		for i, node := range role.Permissions {
			if node.Node == perm.Node {
				role.Permissions[i].State = perm.State
				existed = true
				continue
			}
		}
		if !existed {
			role.Permissions = append(role.Permissions, model.PermissionNode{Node: perm.Node, State: perm.State})
		}
	}

	err = s.repo.UpdateRole(ctx, role)

	if err != nil {
		zap.S().Errorw("error updating role", "error", err)
		return nil, err
	}

	err = s.notif.RoleUpdate(ctx, role, permission2.RoleUpdateMessage_MODIFY)
	if err != nil {
		zap.S().Errorw("error sending role update notification", "error", err)
	}

	return &permission.UpdateRoleResponse{
		Role: role.ToProto(),
	}, nil
}

func (s *permissionService) AddRoleToPlayer(ctx context.Context, req *permission.AddRoleToPlayerRequest) (*permission.AddRoleToPlayerResponse, error) {
	pId, err := uuid.Parse(req.PlayerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid player id %s", req.PlayerId))
	}

	ok, err := s.repo.DoesRoleExist(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}

	if !ok {
		st := status.New(codes.NotFound, "role not found")
		st, _ = st.WithDetails(&permission.AddRoleToPlayerError{ErrorType: permission.AddRoleToPlayerError_ROLE_NOT_FOUND})
		return nil, st.Err()
	}

	err = s.repo.AddRoleToPlayer(ctx, pId, req.RoleId)

	// NOTE: err no documents should never be thrown here because if so, we create a new player with role + default role
	if err != nil {
		if err == repository.AlreadyHasRoleError {
			st := status.New(codes.AlreadyExists, "player already has role")
			st, _ = st.WithDetails(&permission.AddRoleToPlayerError{ErrorType: permission.AddRoleToPlayerError_ALREADY_HAS_ROLE})
			return nil, st.Err()
		}
		return nil, err
	}

	err = s.notif.PlayerRolesUpdate(ctx, pId.String(), req.RoleId, permission2.PlayerRolesUpdateMessage_ADD)
	if err != nil {
		zap.S().Errorw("error sending player roles update", "error", err)
	}

	return &permission.AddRoleToPlayerResponse{}, nil
}

func (s *permissionService) RemoveRoleFromPlayer(ctx context.Context, req *permission.RemoveRoleFromPlayerRequest) (*permission.RemoveRoleFromPlayerResponse, error) {
	pId, err := uuid.Parse(req.PlayerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid player id %s", req.PlayerId))
	}

	err = s.repo.RemoveRoleFromPlayer(ctx, pId, req.RoleId)
	if err != nil {
		if err == mongoDb.ErrNoDocuments {
			st := status.New(codes.NotFound, "player not found")
			st, _ = st.WithDetails(&permission.RemoveRoleFromPlayerError{ErrorType: permission.RemoveRoleFromPlayerError_PLAYER_NOT_FOUND})
			return nil, st.Err()
		} else if err == repository.DoesNotHaveRoleError {
			st := status.New(codes.NotFound, "player does not have role")
			st, _ = st.WithDetails(&permission.RemoveRoleFromPlayerError{ErrorType: permission.RemoveRoleFromPlayerError_DOES_NOT_HAVE_ROLE})
			return nil, st.Err()
		}
		return nil, err
	}

	err = s.notif.PlayerRolesUpdate(ctx, pId.String(), req.RoleId, permission2.PlayerRolesUpdateMessage_REMOVE)
	if err != nil {
		zap.S().Errorw("error sending player roles update", "error", err)
	}

	return &permission.RemoveRoleFromPlayerResponse{}, nil
}

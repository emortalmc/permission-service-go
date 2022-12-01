package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/towerdefence-cc/grpc-api-specs/gen/go/service/permission"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"permission-service-go/mongo/model"
	"permission-service-go/service"
)

const (
	port = 9090
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	permission.RegisterPermissionServiceServer(server, &permissionServer{})

	log.Printf("Starting server on port %d", port)
	if err := server.Serve(lis); err != nil {
		panic(err)
	}
}

type permissionServer struct {
	permission.UnimplementedPermissionServiceServer
}

func (s *permissionServer) GetRoles(context.Context, *emptypb.Empty) (*permission.RolesResponse, error) {
	roles, err := service.GetRoles(context.Background())
	if err != nil {
		log.Print(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &permission.RolesResponse{
		Roles: model.ConvertRoles(roles),
	}, nil
}

func (s *permissionServer) GetPlayerRoles(ctx context.Context, request *permission.PlayerRequest) (*permission.PlayerRolesResponse, error) {
	id, err := uuid.Parse(request.PlayerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	roleIds, err := service.GetPlayerRoleIds(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &permission.PlayerRolesResponse{
		RoleIds: roleIds,
	}, nil
}

func (s *permissionServer) CreateRole(ctx context.Context, request *permission.RoleCreateRequest) (*permission.RoleResponse, error) {
	role, err := service.CreateRole(ctx, request)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return model.ConvertRole(role), nil
}

func (s *permissionServer) UpdateRole(ctx context.Context, request *permission.RoleUpdateRequest) (*permission.RoleResponse, error) {
	role, err := service.UpdateRole(ctx, request)
	if err != nil {
		return nil, err
	}

	return model.ConvertRole(role), nil
}

func (s *permissionServer) AddRoleToPlayer(ctx context.Context, request *permission.AddRoleToPlayerRequest) (*emptypb.Empty, error) {
	playerId, err := uuid.Parse(request.PlayerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = service.AddRoleToPlayer(ctx, playerId, request.RoleId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *permissionServer) RemoveRoleFromPlayer(ctx context.Context, request *permission.RemoveRoleFromPlayerRequest) (*emptypb.Empty, error) {
	playerId, err := uuid.Parse(request.PlayerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = service.RemoveRoleFromPlayer(ctx, playerId, request.RoleId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

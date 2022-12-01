package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/towerdefence-cc/grpc-api-specs/gen/go/service/permission"
	"go.mongodb.org/mongo-driver/bson"
	mongoDb "go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"permission-service-go/mongo"
	"permission-service-go/mongo/model"
)

var (
	roleCollection   = mongo.Database.Collection("roles")
	playerCollection = mongo.Database.Collection("players")
)

func GetRoles(ctx context.Context) ([]*model.Role, error) {

	cursor, err := roleCollection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var mongoResult []model.Role
	err = cursor.All(ctx, &mongoResult)

	slice := make([]*model.Role, len(mongoResult))
	for i := range mongoResult {
		slice[i] = &mongoResult[i]
	}

	return slice, err
}

func GetPlayerRoleIds(ctx context.Context, playerId uuid.UUID) ([]string, error) {
	filter := bson.D{{"_id", playerId}}

	var result model.Player
	err := playerCollection.FindOne(ctx, filter).Decode(&result)

	if err != nil {
		// insert into db if not exists
		if err == mongoDb.ErrNoDocuments {
			_, err := playerCollection.InsertOne(ctx, model.Player{Id: playerId, Roles: []string{"default"}})

			if err != nil {
				return nil, err
			}

			return []string{"default"}, nil
		}
	}

	return result.Roles, err
}

func CreateRole(ctx context.Context, request *permission.RoleCreateRequest) (*model.Role, error) {
	role := model.Role{
		Id:                  request.Id,
		Priority:            request.Priority,
		DisplayPrefix:       *request.DisplayPrefix,
		DisplayNameModifier: *request.DisplayName,
		Permissions:         make([]model.PermissionNode, 0),
	}

	_, err := roleCollection.InsertOne(ctx, role)

	return &role, err
}

func UpdateRole(ctx context.Context, request *permission.RoleUpdateRequest) (*model.Role, error) {
	var role model.Role
	err := roleCollection.FindOne(ctx, bson.D{{"_id", request.Id}}).Decode(&role)

	if err != nil {
		if err == mongoDb.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "Role not found")
		}
		log.Print(err)
		return nil, err
	}

	if request.Priority != nil {
		role.Priority = *request.Priority
	}
	if request.DisplayPrefix != nil {
		role.DisplayPrefix = *request.DisplayPrefix
	}
	if request.DisplayName != nil {
		role.DisplayNameModifier = *request.DisplayName
	}

	for _, perm := range request.UnsetPermissions {
		for i, node := range role.Permissions {
			if node.Node == perm {
				role.Permissions = append(role.Permissions[:i], role.Permissions[i+1:]...)
			}
		}
	}

	for _, perm := range request.SetPermissions {
		var conState string
		if perm.State == permission.PermissionNode_ALLOW {
			conState = model.PermissionStateAllow
		} else {
			conState = model.PermissionStateDeny
		}

		role.Permissions = append(role.Permissions, model.PermissionNode{Node: perm.Node, PermissionState: conState})
	}

	_, err = roleCollection.ReplaceOne(ctx, bson.D{{"_id", request.Id}}, role)

	return &role, err
}

func AddRoleToPlayer(ctx context.Context, playerId uuid.UUID, roleId string) error {
	result, err := playerCollection.UpdateOne(ctx, bson.M{"_id": playerId}, bson.M{"$addToSet": bson.M{"roles": roleId}})

	if err != nil {
		if err == mongoDb.ErrNoDocuments {
			return status.Error(codes.NotFound, "player not found")
		}
		return err
	}

	if result.ModifiedCount == 0 {
		v := errdetails.ErrorInfo{
			Metadata: map[string]string{"custom_cause": "ALREADY_HAS_ROLE"},
		}

		st, _ := status.New(codes.FailedPrecondition, fmt.Sprintf("Player %s already has role %s", playerId, roleId)).WithDetails(&v)
		return st.Err()
	}

	return err
}

func RemoveRoleFromPlayer(ctx context.Context, playerId uuid.UUID, roleId string) error {
	result, err := playerCollection.UpdateOne(ctx, bson.D{{"_id", playerId}}, bson.M{"$pull": bson.M{"roles": roleId}})

	if err != nil {
		if err == mongoDb.ErrNoDocuments {
			return status.Error(codes.NotFound, "player not found")
		}
	}
	if result.ModifiedCount == 0 {
		v := errdetails.ErrorInfo{
			Metadata: map[string]string{"custom_cause": "NO_ROLE"},
		}

		st, _ := status.New(codes.FailedPrecondition, fmt.Sprintf("Player %s does not have role %s", playerId, roleId)).WithDetails(&v)
		return st.Err()
	}

	return err
}

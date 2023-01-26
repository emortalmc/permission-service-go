package model

import (
	protoModel "github.com/emortalmc/proto-specs/gen/go/model/permission"
	"github.com/google/uuid"
)

type Role struct {
	Id            string           `bson:"_id"`
	Priority      uint32           `bson:"priority"`
	DisplayPrefix *string          `bson:"displayPrefix"`
	DisplayName   *string          `bson:"displayName"`
	Permissions   []PermissionNode `bson:"permissions"`
}

func (r *Role) ToProto() *protoModel.Role {
	protoPermissions := make([]*protoModel.PermissionNode, 0)
	for _, p := range r.Permissions {
		protoPermissions = append(protoPermissions, p.ToProto())
	}

	return &protoModel.Role{
		Id:            r.Id,
		Priority:      r.Priority,
		DisplayPrefix: r.DisplayPrefix,
		DisplayName:   r.DisplayName,
		Permissions:   protoPermissions,
	}
}

type PermissionNode struct {
	Node            string                                    `bson:"node"`
	PermissionState protoModel.PermissionNode_PermissionState `bson:"permissionState"`
}

func (p *PermissionNode) ToProto() *protoModel.PermissionNode {
	return &protoModel.PermissionNode{
		Node:  p.Node,
		State: p.PermissionState,
	}
}

// Player does not have an equivalent proto
// as roles are retrieved separately from the player
type Player struct {
	Id    uuid.UUID `bson:"_id"`
	Roles []string  `bson:"roles"`
}

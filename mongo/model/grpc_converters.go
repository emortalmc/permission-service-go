package model

import "github.com/towerdefence-cc/grpc-api-specs/gen/go/service/permission"

func ConvertRole(role *Role) *permission.RoleResponse {
	return &permission.RoleResponse{
		Id:            role.Id,
		Priority:      role.Priority,
		DisplayPrefix: &role.DisplayPrefix,
		DisplayName:   &role.DisplayNameModifier,
		Permissions:   ConvertPermissions(role.Permissions),
	}
}

func ConvertRoles(roles []*Role) []*permission.RoleResponse {
	var result []*permission.RoleResponse

	for _, role := range roles {
		result = append(result, ConvertRole(role))
	}

	return result
}

func ConvertPermission(node PermissionNode) *permission.PermissionNode {
	return &permission.PermissionNode{
		Node:  node.Node,
		State: ConvertPermissionState(node.PermissionState),
	}
}

func ConvertPermissionState(state string) permission.PermissionNode_PermissionState  {
	switch state {
	case "ALLOW":
		return permission.PermissionNode_ALLOW
	case "DENY":
		return permission.PermissionNode_DENY
	default:
		return permission.PermissionNode_DENY
	}
}

func ConvertPermissions(nodes []PermissionNode) []*permission.PermissionNode {
	var result []*permission.PermissionNode

	for _, node := range nodes {
		result = append(result, ConvertPermission(node))
	}

	return result
}

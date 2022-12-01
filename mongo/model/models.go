package model

import "github.com/google/uuid"

type Role struct {
	Id                  string           `bson:"_id"`
	Priority            uint32           `bson:"priority"`
	DisplayPrefix       string           `bson:"displayPrefix"`
	DisplayNameModifier string           `bson:"displayNameModifier"`
	Permissions         []PermissionNode `bson:"permissions"`
}

const (
	PermissionStateAllow = "ALLOW"
	PermissionStateDeny  = "DENY"
)

type PermissionNode struct {
	Node            string `bson:"node"`
	PermissionState string `bson:"permissionState"`
}

type Player struct {
	Id    uuid.UUID `bson:"_id"`
	Roles []string  `bson:"roles"`
}

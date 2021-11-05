package db

import "github.com/diamondburned/arikawa/v3/discord"

// PermissionConfig ...
type PermissionConfig struct {
	BotOwners []discord.UserID `json:"-"`

	User   []PermissionOverride `json:"user"`
	Helper []PermissionOverride `json:"helper"`
	Staff  []PermissionOverride `json:"staff"`
	Owner  []PermissionOverride `json:"owner"`
}

// Level returns the permission level of the given member.
func (p PermissionConfig) Level(m *discord.Member) PermissionLevel {
	for _, o := range p.BotOwners {
		if o == m.User.ID {
			return OwnerLevel
		}
	}

	for _, o := range p.Owner {
		if o.Has(m) {
			return OwnerLevel
		}
	}

	for _, o := range p.Staff {
		if o.Has(m) {
			return StaffLevel
		}
	}

	for _, o := range p.Helper {
		if o.Has(m) {
			return HelperLevel
		}
	}

	for _, o := range p.User {
		if o.Has(m) {
			return UserLevel
		}
	}

	return EveryoneLevel
}

// PermissionOverride ...
type PermissionOverride struct {
	ID   discord.Snowflake `json:"id"`
	Type PermissionType    `json:"type"`
}

// Has ...
func (o PermissionOverride) Has(m *discord.Member) bool {
	if o.Type == UserPermission {
		return m.User.ID == discord.UserID(o.ID)
	}

	for _, r := range m.RoleIDs {
		if r == discord.RoleID(o.ID) {
			return true
		}
	}
	return false
}

// PermissionType is command permission types
type PermissionType int

// Permission type constants
const (
	UserPermission PermissionType = 0
	RolePermission PermissionType = 1
)

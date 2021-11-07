package db

// DefaultPermissions ...
var defaultPermissions = map[string]PermissionLevel{
	"ping":   EveryoneLevel,
	"help":   EveryoneLevel,
	"config": OwnerLevel,
	"app":    StaffLevel,
}

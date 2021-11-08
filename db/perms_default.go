package db

// DefaultPermissions ...
var defaultPermissions = map[string]PermissionLevel{
	"ping":   EveryoneLevel,
	"help":   UserLevel,
	"config": OwnerLevel,
	"app":    StaffLevel,
	"verify": HelperLevel,
	"close":  HelperLevel,
}

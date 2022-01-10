package db

// DefaultPermissions ...
var DefaultPermissions = map[string]PermissionLevel{
	"ping":        EveryoneLevel,
	"help":        UserLevel,
	"config":      OwnerLevel,
	"permissions": OwnerLevel,
	"app":         StaffLevel,
	"verify":      HelperLevel,
	"close":       HelperLevel,
	"deny":        HelperLevel,
	"logs":        HelperLevel,
	"userinfo":    UserLevel,
	"unverified":  StaffLevel,
	"level":       UserLevel,
	"levelcfg":    StaffLevel,
	"restart":     HelperLevel,
}

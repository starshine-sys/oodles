package db

import (
	"errors"
	"fmt"
	"strings"
)

// ConfigOption is a single configuration option definition
type ConfigOption struct {
	Description  string
	Type         ConfigOptionType
	DefaultValue interface{}
	ValidValues  []interface{}
}

// ConfigOptionType is the configuration option's type (string, bool, float, int)
type ConfigOptionType int

// Option type constants
const (
	InvalidOptionType ConfigOptionType = iota
	StringOptionType
	BoolOptionType
	IntOptionType
	FloatOptionType
	SnowflakeOptionType
)

func (t ConfigOptionType) String() string {
	switch t {
	case StringOptionType:
		return "string"
	case BoolOptionType:
		return "boolean"
	case IntOptionType:
		return "integer"
	case FloatOptionType:
		return "float"
	case SnowflakeOptionType:
		return "snowflake/Discord ID"
	default:
		return "unknown type"
	}
}

// Config is a server's configuration.
type Config map[string]interface{}

// Get gets the given configuration setting.
// *This method panics if an invalid option name is given.*
func (c Config) Get(name string) ConfigSetting {
	name = strings.ToLower(name)

	def, ok := ConfigOptions[name]
	if !ok {
		panic("invalid config setting " + name)
	}

	v, ok := c[name]
	if !ok {
		return ConfigSetting{def.DefaultValue}
	}

	return ConfigSetting{v}
}

// Configuration errors
var (
	ErrInvalidConfigOption = errors.New("invalid configuration option")
)

// Set sets the given option name to the given value.
// Passing a nil pointer will delete the value, making it fall back to the default value.
func (c Config) Set(name string, val interface{}) error {
	name = strings.ToLower(name)

	_, ok := ConfigOptions[name]
	if !ok {
		return ErrInvalidConfigOption
	}

	if val == nil {
		delete(c, name)
		return nil
	}

	c[name] = val
	return nil
}

// PermissionLevel is command permission levels
type PermissionLevel int

// Permission level constants
const (
	InvalidLevel  PermissionLevel = 0
	EveryoneLevel PermissionLevel = 1
	UserLevel     PermissionLevel = 2
	HelperLevel   PermissionLevel = 3
	StaffLevel    PermissionLevel = 4
	OwnerLevel    PermissionLevel = 5
	DisabledLevel PermissionLevel = 6
)

func (p PermissionLevel) String() string {
	switch p {
	case EveryoneLevel:
		return "[1] EVERYONE"
	case UserLevel:
		return "[2] USER"
	case HelperLevel:
		return "[3] HELPER"
	case StaffLevel:
		return "[4] STAFF"
	case OwnerLevel:
		return "[5] OWNER"
	case DisabledLevel:
		return "[6] DISABLED"
	default:
		return fmt.Sprintf("[%d] UNKNOWN", int(p))
	}
}

// CommandOverrides is a map of command permission level overrides.
//
// Command permission levels:
// - 1: @everyone
// - 2: normal users
// - 3: chat moderators/helpers
// - 4: staff
// - 5: owner
type CommandOverrides map[string]PermissionLevel

// For returns the permission level for the given command.
func (c CommandOverrides) For(cmd string) PermissionLevel {
	cmd = strings.ToLower(cmd)

	lvl, ok := c[cmd]
	if ok {
		return lvl
	}

	return defaultPermissions[cmd]
}

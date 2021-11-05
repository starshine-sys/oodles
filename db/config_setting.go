package db

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/spf13/cast"
)

// ConfigSetting ...
type ConfigSetting struct {
	val interface{}
}

// ToString returns the ConfigSetting as a string
func (s ConfigSetting) ToString() string {
	return cast.ToString(s.val)
}

// ToBool returns the ConfigSetting as a bool
func (s ConfigSetting) ToBool() bool {
	return cast.ToBool(s.val)
}

// ToInt returns the ConfigSetting as an int64
func (s ConfigSetting) ToInt() int64 {
	return cast.ToInt64(s.val)
}

// ToUint returns the ConfigSetting as a uint64
func (s ConfigSetting) ToUint() uint64 {
	return cast.ToUint64(s.val)
}

// ToFloat returns the ConfigSetting as a float64
func (s ConfigSetting) ToFloat() float64 {
	return cast.ToFloat64(s.val)
}

// ToChannelID returns the ConfigSetting as a discord.ChannelID
func (s ConfigSetting) ToChannelID() discord.ChannelID {
	return discord.ChannelID(s.ToUint())
}

// ToRoleID returns the ConfigSetting as a discord.RoleID
func (s ConfigSetting) ToRoleID() discord.RoleID {
	return discord.RoleID(s.ToUint())
}

// ToUserID returns the ConfigSetting as a discord.UserID
func (s ConfigSetting) ToUserID() discord.UserID {
	return discord.UserID(s.ToUint())
}

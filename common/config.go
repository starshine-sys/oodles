package common

import "github.com/diamondburned/arikawa/v3/discord"

// BotConfig is the bot's configuration file.
// This contains authentication as well as data that is too complicated to store in a key-value format.
type BotConfig struct {
	Token    string `toml:"token"`
	Database string `toml:"database"`

	// Immutable owners, have access to all commands regardless of overrides (except disabled commands)
	Owners  []discord.UserID `toml:"owners"`
	GuildID discord.GuildID  `toml:"guild_id"`

	Backgrounds []Background `toml:"backgrounds"`
}

// Background is a level background.
type Background struct {
	Name        string `toml:"name"`
	Filename    string `toml:"filename"`
	Emoji       string `toml:"emoji"`
	Description string `toml:"description"`
}

package bot

import "github.com/diamondburned/arikawa/v3/discord"

// RootChannel returns the given channel's root channel--either the channel itself, or the parent channel if it's a thread.
func (bot *Bot) RootChannel(id discord.ChannelID) (*discord.Channel, error) {
	ch, err := bot.State.Channel(id)
	if err != nil {
		return nil, err
	}

	if IsThread(ch) {
		return bot.State.Channel(ch.ParentID)
	}

	return ch, nil
}

func IsThread(ch *discord.Channel) bool {
	switch ch.Type {
	case discord.GuildNewsThread, discord.GuildPrivateThread, discord.GuildPublicThread:
		return true
	default:
		return false
	}
}

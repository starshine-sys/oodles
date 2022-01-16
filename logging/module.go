package logging

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/oodles/bot"
)

type Bot struct {
	*bot.Bot

	invites []discord.Invite
	members map[discord.UserID]discord.Member

	invitesMu, membersMu sync.Mutex
}

func Init(bot *bot.Bot) {
	b := &Bot{
		Bot:     bot,
		members: make(map[discord.UserID]discord.Member),
	}

	// backend handlers (ensure correct state)
	b.Router.AddHandler(b.guildCreate)
	b.Router.AddHandler(b.guildMembersChunk)
	b.Router.AddHandler(b.inviteCreate)
	b.Router.AddHandler(b.inviteDelete)

	// actual logging
	b.Router.AddHandler(b.guildMemberAdd)
	b.Router.AddHandler(b.guildMemberRemove)
}

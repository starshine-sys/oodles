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

	ProxiedTriggers   map[discord.MessageID]struct{}
	ProxiedTriggersMu sync.Mutex

	HandledMessages   map[discord.MessageID]struct{}
	HandledMessagesMu sync.Mutex
}

func Init(bot *bot.Bot) {
	b := &Bot{
		Bot:             bot,
		members:         make(map[discord.UserID]discord.Member),
		ProxiedTriggers: make(map[discord.MessageID]struct{}),
		HandledMessages: make(map[discord.MessageID]struct{}),
	}

	// backend handlers (ensure correct state)
	b.Router.AddHandler(b.guildCreate)
	b.Router.AddHandler(b.guildMembersChunk)
	b.Router.AddHandler(b.inviteCreate)
	b.Router.AddHandler(b.inviteDelete)

	// actual logging
	b.Router.AddHandler(b.guildMemberAdd)
	b.Router.AddHandler(b.guildMemberRemove)
	b.Router.AddHandler(b.messageCreate)
	b.Router.AddHandler(b.messageUpdate)
	b.Router.AddHandler(b.messageDelete)
	b.Router.AddHandler(b.bulkMessageDelete)
}

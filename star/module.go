package star

import (
	"context"
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/oodles/bot"
	"github.com/starshine-sys/oodles/star/queries"
)

type Bot struct {
	*bot.Bot

	queries queries.Querier

	mus map[discord.MessageID]*sync.Mutex
	mu  sync.Mutex
}

func Init(b *bot.Bot) {
	bot := &Bot{
		Bot:     b,
		queries: queries.NewQuerier(b.DB),
		mus:     make(map[discord.MessageID]*sync.Mutex),
	}

	bot.Router.AddHandler(bot.reactionAdd)
	bot.Router.AddHandler(bot.reactionRemove)
	bot.Router.AddHandler(bot.reactionRemoveAll)
	bot.Router.AddHandler(bot.reactionRemoveEmoji)
	bot.Router.AddHandler(bot.messageDelete)
}

func (bot *Bot) override(guildID discord.GuildID, channelID discord.ChannelID, categoryID discord.ChannelID) (row queries.ChannelConfigRow, err error) {
	channelHasOverride, err := bot.queries.HasOverride(context.Background(), int64(channelID))
	if err != nil {
		return row, err
	}

	if channelHasOverride {
		return bot.queries.ChannelConfig(context.Background(), int64(channelID), int64(guildID))
	}

	return bot.queries.ChannelConfig(context.Background(), int64(categoryID), int64(guildID))
}

func (bot *Bot) addReaction(message discord.MessageID, user discord.UserID) (int64, error) {
	_, err := bot.queries.AddReaction(context.Background(), int64(user), int64(message))
	if err != nil {
		return 0, err
	}

	return bot.queries.ReactionCount(context.Background(), int64(message))
}

func (bot *Bot) removeReaction(message discord.MessageID, user discord.UserID) (int64, error) {
	_, err := bot.queries.RemoveReaction(context.Background(), int64(user), int64(message))
	if err != nil {
		return 0, err
	}

	return bot.queries.ReactionCount(context.Background(), int64(message))
}

func emojiString(e discord.Emoji) string {
	if e.IsCustom() {
		return e.String()
	}
	return e.Name
}

func (bot *Bot) acquire(id discord.MessageID) func() {
	bot.mu.Lock()
	defer bot.mu.Unlock()

	mu, ok := bot.mus[id]
	if !ok {
		mu = &sync.Mutex{}
		bot.mus[id] = mu
	}
	mu.Lock()
	return mu.Unlock
}

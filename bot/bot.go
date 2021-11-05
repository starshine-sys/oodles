package bot

import (
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/bcr/bot"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

// Bot ...
type Bot struct {
	*bot.Bot

	DB *db.DB
}

// Colour is the embed colour used throughout the bot
const Colour = 0x2A52BE

// Intents are the bot's gateway intents
const Intents = gateway.IntentGuilds | gateway.IntentGuildMembers | gateway.IntentGuildBans | gateway.IntentGuildInvites | gateway.IntentGuildPresences | gateway.IntentGuildMessages | gateway.IntentGuildMessageReactions | gateway.IntentDirectMessages | gateway.IntentDirectMessageReactions

// New ...
func New(conf common.BotConfig) (b *Bot, err error) {
	b = &Bot{}

	b.DB, err = db.New(conf)
	if err != nil {
		return nil, err
	}

	r, err := bcr.NewWithIntents(conf.Token, conf.Owners, nil, Intents)
	if err != nil {
		return nil, err
	}
	b.Bot = bot.NewWithRouter(r)

	b.Router.EmbedColor = Colour
	b.Router.Prefixer = b.DB.Prefixer

	b.Router.AddHandler(b.Router.MessageCreate)
	b.Router.AddHandler(b.Router.InteractionCreate)

	// bot handlers
	b.Router.AddHandler(b.WaitForGuild)

	return b, nil
}

var receivedBotGuild = false

// WaitForGuild ...
func (bot *Bot) WaitForGuild(ev *gateway.GuildCreateEvent) {
	if receivedBotGuild {
		return
	}

	if ev.ID == bot.DB.BotConfig.GuildID {
		receivedBotGuild = true
		common.Log.Infof("Received guild create event for bot guild (%v, %v), bot is ready!", ev.Name, ev.ID)
	}
}

// CheckIfReady ...
func (bot *Bot) CheckIfReady() {
	if !receivedBotGuild {
		common.Log.Warnf("Didn't receive a guild create event for the bot guild (ID %v)! Bot will not function correctly.", bot.DB.BotConfig.GuildID)
	}
}

package bot

import (
	"context"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/bcr/bot"
	bcr2 "github.com/starshine-sys/bcr/v2"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
	"github.com/starshine-sys/pkgo/v2"
)

// Bot ...
type Bot struct {
	*bot.Bot

	Colour discord.Color

	Interactions *bcr2.Router
	State        *state.State
	DB           *db.DB
	Scheduler    *Scheduler
	Checker      *Checker
	PK           *pkgo.Session
}

// Colour is the embed colour used throughout the bot
const Colour = 0x2A52BE

// Intents are the bot's gateway intents
const Intents = gateway.IntentGuilds | gateway.IntentGuildMembers | gateway.IntentGuildBans | gateway.IntentGuildInvites | gateway.IntentGuildPresences | gateway.IntentGuildMessages | gateway.IntentGuildMessageReactions | gateway.IntentDirectMessages | gateway.IntentDirectMessageReactions

// New ...
func New(conf common.BotConfig) (b *Bot, err error) {
	b = &Bot{
		Colour: Colour,
		PK:     pkgo.New(""),
	}

	b.DB, err = db.New(conf)
	if err != nil {
		return nil, err
	}
	b.Checker = &Checker{b.DB, b}

	r, err := bcr.NewWithIntents(conf.Token, conf.Owners, nil, Intents)
	if err != nil {
		return nil, err
	}
	b.Bot = bot.NewWithRouter(r)
	b.Scheduler = NewScheduler(b)
	b.Interactions = bcr2.NewFromShardManager("Bot "+conf.Token, b.Router.ShardManager)

	b.Router.EmbedColor = Colour

	b.Router.AddHandler(b.Router.MessageCreate)
	b.Router.AddHandler(b.interactionCreate)

	// bot handlers
	b.Router.AddHandler(b.WaitForGuild)
	b.Router.AddHandler(b.Ready)

	b.State, _ = b.Router.StateFromGuildID(b.DB.BotConfig.GuildID)

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
	} else {
		common.Log.Warnf("Received guild create event for non-bot guild %v (%v)", ev.Name, ev.ID)
	}
}

// CheckIfReady ...
func (bot *Bot) CheckIfReady() {
	if !receivedBotGuild {
		common.Log.Warnf("Didn't receive a guild create event for the bot guild (ID %v)! Bot will not function correctly.", bot.DB.BotConfig.GuildID)
	}
}

// Prefix only exists because i'm lazy
func (bot *Bot) Prefix() string {
	return bot.DB.Config.Get("prefix").ToString()
}

// Ready ...
func (bot *Bot) Ready(*gateway.ReadyEvent) {
	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	usd := &gateway.UpdatePresenceCommand{
		Status: discord.Status(bot.DB.Config.Get("status").ToString()),
	}

	activity := bot.DB.Config.Get("activity").ToString()
	activityType := bot.DB.Config.Get("activity_type").ToString()

	if activity != "" {
		a := discord.Activity{
			Name: activity,
		}

		switch activityType {
		case "listening":
			a.Type = discord.ListeningActivity
		default:
			a.Type = discord.GameActivity
		}

		usd.Activities = []discord.Activity{a}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	err := s.Gateway().Send(ctx, usd)
	if err != nil {
		common.Log.Errorf("Error setting status: %v", err)
	}
}

func (bot *Bot) interactionCreate(ev *gateway.InteractionCreateEvent) {
	err := bot.Interactions.Execute(ev)
	if err == bcr2.ErrUnknownCommand {
		bot.Router.InteractionCreate(ev)
	} else if err != nil {
		common.Log.Errorf("error in bcr v2 handler: %v", err)
		return
	}
}

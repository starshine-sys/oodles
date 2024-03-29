package applications

import (
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/mozillazg/go-unidecode"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) guildMemberAdd(m *gateway.GuildMemberAddEvent) {
	// don't announce bots joining, and don't announce anything from outside the bot guild
	if m.User.Bot || m.GuildID != bot.DB.BotConfig.GuildID {
		return
	}

	ch := bot.DB.Config.Get("welcome_channel").ToChannelID()
	if !ch.IsValid() {
		return
	}

	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	_, err := s.SendMessage(ch, m.Mention()+" has joined the server!")
	if err != nil {
		common.Log.Errorf("Error sending message: %v", err)
	}
}

func (bot *Bot) guildMemberRemove(ev *gateway.GuildMemberRemoveEvent) {
	if ev.User.Bot || ev.GuildID != bot.DB.BotConfig.GuildID {
		return
	}

	app, err := bot.DB.UserApplication(ev.User.ID)
	if err != nil {
		// no app
		return
	}

	bot.deniedMu.RLock()
	defer bot.deniedMu.RUnlock()
	if _, ok := bot.denied[ev.User.ID]; ok {
		return
	}

	if app.Closed || (app.Verified != nil && *app.Verified) {
		return
	}

	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	_, err = s.SendMessage(app.ChannelID, fmt.Sprintf("📤 %v/%v has left the server.", ev.User.Tag(), ev.User.Mention()))
	if err != nil {
		common.Log.Errorf("Error sending message: %v", err)
	}

	newCat := discord.ChannelID(bot.DB.Config.Get("finished_application_category").ToSnowflake())
	if !newCat.IsValid() {
		newCat = discord.ChannelID(bot.DB.Config.Get("application_category").ToSnowflake())
	}

	var overwrites *[]discord.Overwrite
	cat, err := s.Channel(newCat)
	if err == nil {
		overwrites = &cat.Overwrites
	}

	err = s.ModifyChannel(app.ChannelID, api.ModifyChannelData{
		CategoryID:     newCat,
		Overwrites:     overwrites,
		Name:           "📤-app-" + unidecode.Unidecode(ev.User.Username),
		AuditLogReason: "User left server before application was completed",
	})
	if err != nil {
		common.Log.Errorf("Error updating channel title: %v", err)
	}

	if app.ScheduledEventID != nil {
		err = bot.Scheduler.Remove(*app.ScheduledEventID)
		if err != nil {
			bot.SendError("Error removing schedled timeout message for app %v: %v", app.ID, err)
		}
	}

	_, err = bot.createTranscript(s, app)
	if common.IsOodlesError(err) {
		_, err = s.SendMessage(app.ChannelID, fmt.Sprintf("❌ %v", err))
		if err != nil {
			common.Log.Errorf("Error sending message: %v", err)
		}
	} else if err != nil {
		common.Log.Errorf("Error saving transcript: %v", err)
		_, err = s.SendMessage(app.ChannelID, fmt.Sprintf("I wasn't able to save a transcript:\n> %v", err))
		if err != nil {
			common.Log.Errorf("Error sending message: %v", err)
		}
		return
	}

	eventID, err := bot.Scheduler.Add(
		time.Now().Add(ScheduledCloseTime), &scheduledClose{ChannelID: app.ChannelID},
	)
	if err == nil {
		if err := bot.DB.SetCloseID(app.ID, eventID); err != nil {
			common.Log.Errorf("error setting scheduled close id for app %v: %v", app.ID, err)
		}
	} else {
		common.Log.Errorf("error adding scheduled close for app %v: %v", app.ID, err)
	}
}

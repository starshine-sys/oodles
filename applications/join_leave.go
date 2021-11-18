package applications

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/mozillazg/go-unidecode"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) guildMemberAdd(m *gateway.GuildMemberAddEvent) {
	ch := bot.DB.Config.Get("welcome_channel").ToChannelID()
	if !ch.IsValid() {
		return
	}

	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	s.SendMessage(ch, m.Mention()+" has joined the server!")
}

func (bot *Bot) guildMemberRemove(ev *gateway.GuildMemberRemoveEvent) {
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

	err = s.ModifyChannel(app.ChannelID, api.ModifyChannelData{
		Name:           "📤-app-" + unidecode.Unidecode(ev.User.Username),
		AuditLogReason: "User left server before application was completed",
	})

	_, err = bot.createTranscript(s, app)
	if common.IsOodlesError(err) {
		s.SendMessage(app.ChannelID, fmt.Sprintf("❌ %v", err))
	} else if err != nil {
		common.Log.Errorf("Error saving transcript: %v", err)
		s.SendMessage(app.ChannelID, fmt.Sprintf("I wasn't able to save a transcript:\n> %v", err))
		return
	}
}
package applications

import (
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/mozillazg/go-unidecode"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) completeApp(app *db.Application, m *gateway.MessageCreateEvent) error {
	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	err := bot.DB.CompleteApp(app.ID)
	if err != nil {
		bot.SendError("Error setting app %v as complete: %v", app.ID, err)
		return err
	}

	msg := bot.DB.Config.Get("application_finished_message").ToString()

	// give time for pk to proxy
	time.Sleep(waitTime)
	err = bot.sendInterviewMessage(app, msg)
	if err != nil {
		bot.SendError("Error sending finished message to app %v: %v", app.ID, err)
		return err
	}

	discussion := bot.DB.Config.Get("discussion_channel").ToChannelID()
	if discussion.IsValid() {
		msg, err := s.SendMessage(discussion, fmt.Sprintf("%v (%v) has finished their application! What do you think?", app.UserID.Mention(), app.ChannelID.Mention()))
		if err == nil {
			go func() {
				for _, e := range []discord.APIEmoji{"✅", "❌", "🤔"} {
					s.React(msg.ChannelID, msg.ID, e)
				}
			}()
		}
	}

	return s.ModifyChannel(app.ChannelID, api.ModifyChannelData{
		Name:           "✅-app-" + unidecode.Unidecode(m.Author.Username),
		AuditLogReason: "Completed application, waiting for followup",
	})
}

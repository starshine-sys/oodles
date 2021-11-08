package bot

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/google/uuid"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
	"go.uber.org/zap"
)

// Report sends an error message to the log channel.
func (bot *Bot) Report(ctx *bcr.Context, err error) error {
	id := uuid.New()

	common.Log.Desugar().Error(
		"Error in command",
		zap.Strings("command", ctx.FullCommandPath),
		zap.Stringer("error_id", id),
		zap.Error(err),
		zap.StackSkip("stack", 1),
	)

	if bot.DB.BotConfig.LogChannel.IsValid() {
		if !receivedBotGuild {
			common.Log.Warnf("Encountered error before receiving bot guild, not logging to log channel!")
		} else {
			_, e := ctx.State.SendMessage(bot.DB.BotConfig.LogChannel, "", discord.Embed{
				Title:       "Error in command",
				Description: fmt.Sprintf("```%v```", err),
				Color:       bcr.ColourRed,
				Fields: []discord.EmbedField{
					{
						Name:  "Error code",
						Value: fmt.Sprintf("`%s`", id),
					},
					{
						Name:  "User/channel",
						Value: fmt.Sprintf("**User:** %v (%v/%v)\n**Channel:** %v (%v)", ctx.Author.Tag(), ctx.Author.Mention(), ctx.Author.ID, ctx.Channel.Mention(), ctx.Channel.ID),
					},
				},
				Timestamp: discord.NowTimestamp(),
			})
			if e != nil {
				common.Log.Errorf("Error sending error report: %v", e)
			}
		}
	}

	return ctx.SendX(fmt.Sprintf("Error code: `%s`", id),
		discord.Embed{
			Title:       "Internal error occurred",
			Description: fmt.Sprintf("Please report the error code above to the devs! (For example, by DMing %v)", bot.Router.Bot.Username),
			Color:       bcr.ColourRed,
		})
}

// SendLog ...
func (bot *Bot) SendLog(tmpl string, args ...interface{}) {
	bot.sendLog(false, tmpl, args...)
}

// SendError ...
func (bot *Bot) SendError(tmpl string, args ...interface{}) {
	bot.sendLog(true, tmpl, args...)
}

func (bot *Bot) sendLog(error bool, tmpl string, args ...interface{}) {
	if error {
		common.Log.Errorf(tmpl, args...)
	} else {
		common.Log.Infof(tmpl, args...)
	}

	if !bot.DB.BotConfig.LogChannel.IsValid() {
		return
	}

	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	_, err := s.SendMessageComplex(bot.DB.BotConfig.LogChannel, api.SendMessageData{
		Content: fmt.Sprintf(tmpl, args...),
	})
	if err != nil {
		common.Log.Errorf("Error sending error report: %v", err)
	}
}

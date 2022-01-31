package applications

import (
	"context"
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/oodles/bot"
	"github.com/starshine-sys/oodles/common"
)

type timeout struct {
	ChannelID discord.ChannelID `json:"channel_id"`
	UserID    discord.UserID    `json:"user_id"`
}

func (dat *timeout) Execute(ctx context.Context, id int64, bot *bot.Bot) error {
	common.Log.Infof("app in channel %v timed out, sending timeout message", dat.ChannelID)

	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	chID := bot.DB.Config.Get("discussion_channel").ToChannelID()
	if !chID.IsValid() {
		return nil
	}

	_, err := s.SendMessage(chID,
		fmt.Sprintf("%v (%v)'s application timed out!", dat.UserID.Mention(), dat.ChannelID.Mention()))
	return err
}

func (dat *timeout) Offset() time.Duration { return time.Minute }

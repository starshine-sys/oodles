package reminders

import (
	"context"
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/bot"
	"github.com/starshine-sys/oodles/common"
)

type reminder struct {
	ID      int64     `json:"-"`
	Expires time.Time `json:"-"`

	UserID       discord.UserID `json:"user_id"`
	ReminderText string         `json:"text"`
	SetTime      time.Time      `json:"set_time"`

	GuildID   discord.GuildID   `json:"guild_id"`
	ChannelID discord.ChannelID `json:"channel_id"`
	MessageID discord.MessageID `json:"message_id"`
}

func (dat *reminder) Execute(ctx context.Context, id int64, bot *bot.Bot) (err error) {
	common.Log.Debugf("Executing reminder #%v", id)

	linkGuild := dat.GuildID.String()
	if !dat.GuildID.IsValid() {
		linkGuild = "@me"
	}

	s, _ := bot.Router.StateFromGuildID(dat.GuildID)

	shouldDM := true
	if dat.GuildID.IsValid() {
		// this is Ugly™ but it works
		// basically we need to get All of them to check perms
		m, err := s.Member(dat.GuildID, dat.UserID)
		if err == nil {
			g, err := s.Guild(dat.GuildID)
			if err == nil {
				ch, err := s.Channel(dat.ChannelID)
				if err == nil {
					perms := discord.CalcOverwrites(*g, *ch, *m)
					if perms.Has(discord.PermissionSendMessages | discord.PermissionViewChannel) {
						shouldDM = false
					}
				}
			}
		}
	}

	str := fmt.Sprintf("%v: %v (%v)", dat.UserID.Mention(), dat.ReminderText, bcr.HumanizeTime(bcr.DurationPrecisionSeconds, dat.SetTime))

	switch shouldDM {
	case false:
		_, err = s.SendMessageComplex(dat.ChannelID, api.SendMessageData{
			Content: str,
			Components: discord.Components(&discord.ButtonComponent{
				Label: "Jump to message",
				Style: discord.LinkButtonStyle(
					fmt.Sprintf("https://discord.com/channels/%v/%v/%v", linkGuild, dat.ChannelID, dat.MessageID),
				),
			}),
		})
		if err == nil {
			return nil
		}

		fallthrough
	default:
		str = fmt.Sprintf("%v (%v)", dat.ReminderText, bcr.HumanizeTime(bcr.DurationPrecisionSeconds, dat.SetTime))

		ch, err := s.CreatePrivateChannel(dat.UserID)
		if err != nil {
			common.Log.Errorf("Error sending reminder %v: %v", id, err)
			return err
		}

		_, err = s.SendMessageComplex(ch.ID, api.SendMessageData{
			Content: str,
			Components: discord.Components(&discord.ButtonComponent{
				Label: "Jump to message",
				Style: discord.LinkButtonStyle(
					fmt.Sprintf("https://discord.com/channels/%v/%v/%v", linkGuild, dat.ChannelID, dat.MessageID),
				),
			}),
		})
		if err != nil {
			common.Log.Errorf("Error sending reminder %v: %v", id, err)
		}
		return err
	}
}

func (*reminder) Offset() time.Duration { return time.Minute }

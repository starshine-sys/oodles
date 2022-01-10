package applications

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/starshine-sys/dischtml"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) createTranscript(s *state.State, app *db.Application) (*discord.Message, error) {
	outcome := "User left the server"
	if app.Verified != nil {
		if *app.Verified {
			outcome = "User was verified"
		} else {
			outcome = "User was denied"
		}
	}

	msg, err := s.SendMessage(app.ChannelID, fmt.Sprintf("Application complete! %v.", outcome))
	if err != nil {
		return nil, err
	}

	tch := bot.DB.Config.Get("transcript_channel").ToChannelID()
	if !tch.IsValid() {
		return nil, common.Error("There is no transcript channel set, can't create a transcript!")
	}

	g, err := s.Guild(bot.DB.BotConfig.GuildID)
	if err != nil {
		return nil, err
	}

	chs, err := s.Channels(g.ID)
	if err != nil {
		return nil, err
	}

	rls, err := s.Roles(g.ID)
	if err != nil {
		return nil, err
	}

	members, err := s.Members(g.ID)
	if err != nil {
		return nil, err
	}

	conv := dischtml.Converter{
		Guild:    *g,
		Channels: chs,
		Roles:    rls,
		Members:  members,
	}

	msgs, err := s.Messages(app.ChannelID, 500)
	if err != nil {
		return nil, err
	}

	for _, m := range msgs {
		for _, u := range m.Mentions {
			conv.Users = append(conv.Users, u.User)
		}
	}

	str, err := conv.ConvertHTML(msgs)
	if err != nil {
		return nil, err
	}

	var ch discord.Channel
	for _, c := range chs {
		if c.ID == app.ChannelID {
			ch = c
			break
		}
	}

	html, err := dischtml.Wrap(*g, ch, str, len(msgs))
	if err != nil {
		return nil, err
	}

	u, err := s.User(app.UserID)
	if err != nil {
		u = &discord.User{
			Username:      "unknown",
			Discriminator: "0000",
			ID:            app.UserID,
		}
	}

	e := discord.Embed{
		Author: &discord.EmbedAuthor{
			Name: u.Tag(),
			Icon: u.AvatarURLWithType(discord.PNGImage),
		},
		Title: "Application transcript",
		Fields: []discord.EmbedField{
			{
				Name:   "Outcome",
				Value:  outcome,
				Inline: true,
			},
			{
				Name:   "User",
				Value:  fmt.Sprintf("%v\n%v\n%v", u.Mention(), u.Tag(), u.ID),
				Inline: true,
			},
			{
				Name:   "Message count",
				Value:  strconv.Itoa(len(msgs)),
				Inline: true,
			},
			{
				Name:   "Opened at",
				Value:  fmt.Sprintf("<t:%v>", app.ID.Time().UTC().Unix()),
				Inline: true,
			},
		},
		Color:     bot.Colour,
		Timestamp: discord.NowTimestamp(),
		Footer: &discord.EmbedFooter{
			Text: "Application ID: " + app.ID.String(),
		},
	}

	if app.TrackID != nil {
		track, err := bot.DB.ApplicationTrack(*app.TrackID)
		if err == nil {
			e.Fields = append(e.Fields, discord.EmbedField{
				Name:   "Track",
				Value:  track.Name,
				Inline: true,
			})
		}
	}

	if app.Moderator != nil {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:   "Moderator",
			Value:  app.Moderator.Mention(),
			Inline: true,
		})
	}

	if app.Verified != nil && !*app.Verified {
		reason := "No reason specified"
		if app.DenyReason != nil {
			reason = *app.DenyReason
		}

		if len(reason) >= 1024 {
			reason = reason[:1020] + "..."
		}

		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "Denial reason",
			Value: reason,
		})
	}

	logMsg, err := s.SendMessageComplex(tch, api.SendMessageData{
		Embeds: []discord.Embed{e},
		Files: []sendpart.File{{
			Name:   "transcript.html",
			Reader: strings.NewReader(html),
		}},
	})
	if err != nil {
		return nil, err
	}

	err = bot.DB.SetTranscript(app.ID, tch, logMsg.ID)
	if err != nil {
		bot.SendError("Error saving transcript in database: %v", err)
	}

	_, err = s.EditMessage(app.ChannelID, msg.ID, fmt.Sprintf("Application complete! %v.\nTranscript link: https://discord.com/channels/%v/%v/%v", outcome, bot.DB.BotConfig.GuildID, tch, logMsg.ID))
	if err != nil {
		common.Log.Errorf("Error sending confirmation message: %v", err)
	}
	return msg, nil
}

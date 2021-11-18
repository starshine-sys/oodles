package applications

import (
	"fmt"
	"strconv"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) logs(ctx *bcr.Context) (err error) {
	u, err := ctx.ParseUser(ctx.Args[0])
	if err != nil {
		return ctx.SendX("User not found.")
	}

	apps, err := bot.DB.AllUserApplications(u.ID)
	if err != nil {
		bot.SendError("Error fetching apps for %v: %v", u.ID, err)
		return ctx.SendfX("There was an error fetching applications!")
	}

	if len(apps) == 0 {
		return ctx.SendX("That user has no applications.")
	}

	var e []discord.Embed
	m := map[discord.UserID]*discord.User{}
	for i, app := range apps {
		e = append(e, bot.appEmbed(ctx, m, i, len(apps), app))
	}

	_, _, err = ctx.ButtonPages(e, 15*time.Minute)
	return err
}

func (bot *Bot) appEmbed(ctx *bcr.Context, m map[discord.UserID]*discord.User, i, num int, app db.Application) discord.Embed {
	var err error

	e := discord.Embed{
		Color: bot.Colour,
		Fields: []discord.EmbedField{{
			Name:   "User",
			Value:  app.UserID.Mention(),
			Inline: true,
		}},
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("ID: %v | Page %v of %v", app.ID, i+1, num),
		},
	}

	u, ok := m[app.UserID]
	if !ok {
		u, err = ctx.State.User(app.UserID)
		if err == nil {
			m[u.ID] = u
		}
	}

	if u != nil {
		e.Author = &discord.EmbedAuthor{
			Name: u.Tag(),
			Icon: u.AvatarURLWithType(discord.PNGImage),
		}
	} else {
		e.Author = &discord.EmbedAuthor{
			Name: "unknown#0000 (" + app.UserID.String() + ")",
		}
	}

	if app.TrackID != nil {
		track, err := bot.DB.ApplicationTrack(*app.TrackID)
		if err == nil {
			e.Fields = append(e.Fields, discord.EmbedField{
				Name:   "Track",
				Value:  track.Name,
				Inline: true,
			})
		} else {
			bot.SendError("Error getting application track %v: %v", *app.TrackID, err)
			e.Fields = append(e.Fields, discord.EmbedField{
				Name:   "Track",
				Value:  "Unknown track " + strconv.FormatInt(*app.TrackID, 10),
				Inline: true,
			})
		}
	}

	e.Fields = append(e.Fields, discord.EmbedField{
		Name:   "Opened",
		Value:  fmt.Sprintf("<t:%v>", app.Opened.Unix()),
		Inline: true,
	})

	if !app.Closed {
		e.Description = "**Note: application is still open, no transcript!**"
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:   "Channel",
			Value:  app.ChannelID.Mention(),
			Inline: true,
		})
	} else {
		if app.TranscriptChannel != nil && app.TranscriptMessage != nil {
			e.Description = fmt.Sprintf("[Link to transcript](https://discord.com/channels/%v/%v/%v)", bot.DB.BotConfig.GuildID, *app.TranscriptChannel, *app.TranscriptMessage)
		}

		if app.ClosedTime != nil {
			e.Fields = append(e.Fields, discord.EmbedField{
				Name:   "Closed",
				Value:  fmt.Sprintf("<t:%v>", app.ClosedTime.Unix()),
				Inline: true,
			})
		}
	}

	e.Fields = append(e.Fields, discord.EmbedField{
		Name:   "Completed interview",
		Value:  strconv.FormatBool(app.Completed),
		Inline: true,
	})

	if app.Verified != nil {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:   "Verified",
			Value:  strconv.FormatBool(*app.Verified),
			Inline: true,
		})

		if app.Moderator != nil {
			e.Fields = append(e.Fields, discord.EmbedField{
				Name:   "Moderator",
				Value:  app.Moderator.Mention(),
				Inline: true,
			})
		}

		if !*app.Verified {
			reason := "No reason given"
			if app.DenyReason != nil && *app.DenyReason != "" {
				reason = *app.DenyReason
			}

			e.Fields = append(e.Fields, discord.EmbedField{
				Name:   "Denial reason",
				Value:  reason,
				Inline: false,
			})
		}
	}

	return e
}

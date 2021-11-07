package meta

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/dustin/go-humanize"
	"github.com/dustin/go-humanize/english"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) listAppTracks(ctx *bcr.Context) (err error) {
	tracks, err := bot.DB.ApplicationTracks()
	if err != nil {
		return bot.Report(ctx, err)
	}

	switch len(tracks) {
	case 0:
		return ctx.SendfX("There are no application tracks. Please create some first (with `%vapp track create`)", bot.Prefix())
	default:
		e := discord.Embed{
			Title:       "Application tracks",
			Description: fmt.Sprintf("The following application tracks are available.\nTo see the list of questions in each, use `%vapp questions`.\n\n", bot.Prefix()),
			Color:       bot.Colour,
		}

		for _, t := range tracks {
			e.Description += fmt.Sprintf("%d. %s (%s)\n", t.ID, t.Name, t.Emoji())
		}

		return ctx.SendX("", e)
	}
}

func (bot *Bot) createAppTrack(ctx *bcr.Context) (err error) {
	name := ctx.Args[0]
	desc := ctx.Args[1]
	emoji := ctx.Args[2]

	t, err := bot.DB.AddApplicationTrack(db.ApplicationTrack{
		Name:        name,
		Description: desc,
		RawEmoji:    emoji,
	})
	if err != nil {
		return ctx.SendfX("Error adding track:\n> %v", err)
	}

	return ctx.SendfX("Added track **%v**, with ID %v and emoji %s", t.Name, t.ID, t.Emoji())
}

func (bot *Bot) addQuestion(ctx *bcr.Context) (err error) {
	if len(ctx.Args) < 2 {
		return ctx.SendX("You need to give a track ID and questions!")
	}

	track := ctx.Args[0]
	trackID, err := strconv.ParseInt(track, 10, 64)
	if err != nil {
		return ctx.SendfX("%v is not a valid number.", track)
	}

	rawQuestions := strings.TrimSpace(strings.TrimPrefix(ctx.RawArgs, track))
	if rawQuestions == ctx.RawArgs {
		rawQuestions = strings.Join(ctx.Args[1:], " ")
	}
	var questions []string
	if !strings.Contains(rawQuestions, "|") {
		questions = []string{rawQuestions}
	} else {
		questions = strings.Split(rawQuestions, "|")
	}
	for i := range questions {
		questions[i] = strings.TrimSpace(questions[i])

		if len(questions[i]) >= 2000 {
			return ctx.SendfX("Sorry, the %v question you gave is too long (maximum 2000 characters). This is a Discord limitation, sorry!", humanize.Ordinal(i+1))
		}
	}

	for i, q := range questions {
		err = bot.DB.AddQuestion(trackID, q)
		if err != nil {
			return ctx.SendfX("Error adding question %v:\n> %v", i+1, err)
		}
	}

	return ctx.SendfX("Success, added %v!", english.Plural(len(questions), "question", ""))
}

func (bot *Bot) listQuestions(ctx *bcr.Context) (err error) {
	tracks, err := bot.DB.ApplicationTracks()
	if err != nil {
		return bot.Report(ctx, err)
	}

	switch len(tracks) {
	case 0:
		return ctx.SendfX("There are no application tracks. Please create some first (with `%vapp track create`)", bot.Prefix())
	case 1:
		qs, err := bot.DB.Questions(tracks[0].ID)
		if err != nil {
			return bot.Report(ctx, err)
		}

		e := discord.Embed{
			Title:       "Application tracks",
			Description: fmt.Sprintf("There is only one application track, %v (%s, ID %d).\nThe questions in that track are shown below:\n\n", tracks[0].Name, tracks[0].Emoji(), tracks[0].ID),
			Color:       bot.Colour,
		}

		if len(qs) == 0 {
			e.Description += "(no questions yet!)"
		} else {
			for _, q := range qs {
				e.Description += fmt.Sprintf("**%d.** %s\n", q.Index, q.Question)
			}
		}

		return ctx.SendX("", e)
	default:
		e := discord.Embed{
			Title:       "Application tracks",
			Description: "The following application tracks are available:\n",
			Color:       bot.Colour,
		}

		sel := discord.SelectComponent{
			CustomID:    "question_select",
			Placeholder: "Which track's questions do you want to view?",
			MinValues:   option.NewInt(1),
			MaxValues:   1,
		}

		for _, t := range tracks {
			e.Description += fmt.Sprintf("%d. %s (%s)\n", t.ID, t.Name, t.Emoji())
			sel.Options = append(sel.Options, discord.SelectComponentOption{
				Label:       t.Name,
				Value:       t.Name,
				Description: t.Description,
				Emoji: &discord.ButtonEmoji{
					Name:     t.Emoji().Name,
					ID:       t.Emoji().ID,
					Animated: t.Emoji().Animated,
				},
			})
		}

		msg, err := ctx.State.SendMessageComplex(ctx.Message.ChannelID, api.SendMessageData{
			Embeds: []discord.Embed{e},
			Components: []discord.Component{&discord.ActionRowComponent{
				Components: []discord.Component{&sel},
			}},
		})
		if err != nil {
			return err
		}

		resp, timeout := bot.WaitForSelect(ctx, msg.ID, "question_select", 5*time.Minute)
		if timeout {
			return nil
		}

		var questions []db.AppQuestion
		for _, t := range tracks {
			if t.Name == resp.Values[0] {
				questions, err = bot.DB.Questions(t.ID)
				break
			}
		}
		if err != nil {
			common.Log.Errorf("Error fetching questions: %v", err)
			return err
		}

		e = discord.Embed{
			Title: "Questions for " + resp.Values[0],
			Color: bot.Colour,
		}

		if len(questions) == 0 {
			e.Description = "(no questions)"
		} else {
			for _, q := range questions {
				e.Description += fmt.Sprintf("**%d.** %s\n", q.Index, q.Question)
			}
		}

		return ctx.State.RespondInteraction(resp.ID, resp.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{e},
			},
		})
	}
}

// SelectResponse ...
type SelectResponse struct {
	ID     discord.InteractionID
	Token  string
	Values []string
}

// WaitForSelect ...
func (bot *Bot) WaitForSelect(ctx *bcr.Context, msgID discord.MessageID, customID string, dur time.Duration) (resp SelectResponse, timeout bool) {
	c, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()

	v := ctx.State.WaitFor(c, func(v interface{}) bool {
		ev, ok := v.(*gateway.InteractionCreateEvent)
		if !ok {
			return false
		}

		if ev.Message == nil || ev.Message.ID != msgID {
			return false
		}

		u := ev.User
		if u == nil {
			u = &ev.Member.User
		}

		if u.ID != ctx.Author.ID {
			return false
		}

		data, ok := ev.Data.(*discord.ComponentInteractionData)
		if !ok {
			return false
		}

		return data.CustomID == customID
	})

	if v == nil {
		return resp, true
	}

	ev := v.(*gateway.InteractionCreateEvent)
	resp = SelectResponse{
		ID:    ev.ID,
		Token: ev.Token,
	}

	data := ev.Data.(*discord.ComponentInteractionData)
	resp.Values = data.Values

	return
}

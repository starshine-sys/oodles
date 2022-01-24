package levels

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) setBackground(ctx *bcr.Context) (err error) {
	if ctx.RawArgs == "random" {
		_, err = bot.DB.Exec(context.Background(), "update levels set background = null where guild_id = $1 and user_id = $2", ctx.Message.GuildID, ctx.Author.ID)
		if err != nil {
			return bot.Report(ctx, err)
		}
		return ctx.SendX("Background preference cleared! You'll now get a random background every time you pull up your level card.")
	}

	lbs, err := bot.backgroundMetadata()
	if err != nil {
		return bot.Report(ctx, err)
	}

	if len(lbs) == 0 {
		return ctx.SendX("Looks like there's no backgrounds available.")
	}

	var btns []*discord.ButtonComponent
	for _, lb := range lbs {
		emoji := &discord.ComponentEmoji{
			Name: lb.EmojiName,
		}
		if lb.EmojiID != nil {
			emoji.ID = *lb.EmojiID
		}

		btns = append(btns, &discord.ButtonComponent{
			Style:    discord.SecondaryButtonStyle(),
			CustomID: discord.ComponentID(fmt.Sprintf("levelbg:%d", lb.ID)),
			Emoji:    emoji,
		})
	}

	btns = append(btns, &discord.ButtonComponent{
		Style:    discord.SecondaryButtonStyle(),
		CustomID: "levelbg:random",
		Emoji: &discord.ComponentEmoji{
			Name: "🎲",
		},
	})

	_, err = ctx.SendComponents(
		buttonsToRows(btns),
		"Click one of the buttons below to choose a level background (check the pins in <#847385848060182528> for options)!",
	)
	return
}

func buttonsToRows(buttons []*discord.ButtonComponent) (rows discord.ContainerComponents) {
	actionRow := &discord.ActionRowComponent{}
	count := 0
	for _, btn := range buttons {
		if count >= 5 {
			rows = append(rows, actionRow)
			actionRow = &discord.ActionRowComponent{}
			count = 0
		}
		*actionRow = append(*actionRow, btn)
		count++
	}

	if len(*actionRow) > 0 {
		rows = append(rows, actionRow)
	}
	return rows
}

func (bot *Bot) chooseBackground(ev *gateway.InteractionCreateEvent) {
	data, ok := ev.Data.(*discord.ButtonInteraction)
	if !ok {
		return
	}

	if data.CustomID == "levelbg:random" {
		_, err := bot.DB.Exec(context.Background(), "update levels set background = null where guild_id = $1 and user_id = $2", ev.GuildID, ev.SenderID())
		if err != nil {
			common.Log.Errorf("Error updating level background: %v", err)
			bot.respond(ev, "Internal error occurred!")
			return
		}

		bot.respond(ev, "Background preference cleared! You'll now get a random background every time you pull up your level card.")
		return
	}

	if strings.HasPrefix(string(data.CustomID), "levelbg:") {
		id, err := strconv.ParseInt(strings.TrimPrefix(string(data.CustomID), "levelbg:"), 10, 64)
		if err != nil {
			bot.respond(ev, "That doesn't seem to be a valid background. Try running the command again?")
			return
		}

		if !bot.bgExists(id) {
			bot.respond(ev, "That doesn't seem to be a valid background. Try running the command again?")
			return
		}

		_, err = bot.DB.Exec(context.Background(), "update levels set background = $3 where guild_id = $1 and user_id = $2", ev.GuildID, ev.SenderID(), id)
		if err != nil {
			common.Log.Errorf("Error updating level background: %v", err)
			bot.respond(ev, "Internal error occurred!")
			return
		}

		bg, err := bot.background(id)
		if err != nil {
			common.Log.Errorf("Error getting level background %v: %v", id, err)
			bot.respond(ev, "Level background changed!")
			return
		}

		bot.respond(ev, "Level background changed to %v!", bg.Source)
	}
}

func (bot *Bot) respond(ev *gateway.InteractionCreateEvent, tmpl string, args ...interface{}) {
	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	err := s.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(fmt.Sprintf(tmpl, args...)),
			Flags:   api.EphemeralResponse,
			AllowedMentions: &api.AllowedMentions{
				Parse: []api.AllowedMentionType{},
			},
		},
	})
	if err != nil {
		common.Log.Errorf("Error responding to interaction: %v", err)
	}
}

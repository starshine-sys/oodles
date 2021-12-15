package levels

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) configShow(ctx *bcr.Context) (err error) {
	gc, err := bot.getGuildConfig(ctx.Guild.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	desc := fmt.Sprintf(`**Levels enabled:** %v
**DM on reward:** %v
**Time between XP:** %v
**Reward log:** %v
**Nolevels log:** %v`, gc.LevelsEnabled, gc.DMOnReward, gc.BetweenXP, bot.mentionOrNone(gc.RewardLog), bot.mentionOrNone(gc.NolevelsLog))

	e := discord.Embed{
		Color:       bot.Colour,
		Title:       "Configuration for " + ctx.Guild.Name,
		Description: desc,
	}

	f := discord.EmbedField{
		Name:   "Blacklisted channels",
		Inline: true,
	}
	if len(gc.BlockedChannels) == 0 {
		f.Value = "None"
	} else {
		for _, ch := range gc.BlockedChannels {
			f.Value += discord.ChannelID(ch).Mention() + "\n"
		}
	}
	e.Fields = append(e.Fields, f)

	f = discord.EmbedField{
		Name:   "Blacklisted categories",
		Inline: true,
	}
	if len(gc.BlockedCategories) == 0 {
		f.Value = "None"
	} else {
		for _, ch := range gc.BlockedCategories {
			f.Value += discord.ChannelID(ch).Mention() + "\n"
		}
	}
	e.Fields = append(e.Fields, f)

	f = discord.EmbedField{
		Name:   "Blacklisted roles",
		Inline: true,
	}
	if len(gc.BlockedRoles) == 0 {
		f.Value = "None"
	} else {
		for _, ch := range gc.BlockedRoles {
			f.Value += discord.RoleID(ch).Mention() + "\n"
		}
	}
	e.Fields = append(e.Fields, f)

	rewards, err := bot.getAllRewards(ctx.Guild.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	fieldVal := ""
	for _, reward := range rewards {
		fieldVal += fmt.Sprintf("%v: %v\n", reward.Level, reward.RoleReward.Mention())
	}
	if fieldVal != "" {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "Level rewards",
			Value: fieldVal,
		})
	}

	if gc.RewardText != "" {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "Reward text",
			Value: "```md\n" + gc.RewardText + "\n```",
		})
	}

	return ctx.SendX("", e)
}

func (bot *Bot) mentionOrNone(id discord.ChannelID) string {
	if id.IsValid() {
		return id.Mention()
	}
	return "None"
}

func (bot *Bot) setConfig(ctx *bcr.Context) (err error) {
	switch strings.ToLower(ctx.Args[0]) {
	case "enable":
		b, err := strconv.ParseBool(ctx.Args[1])
		if err != nil {
			_, err = ctx.Replyc(bcr.ColourRed, "`%v` is not a valid setting for `enable` (true or false)", ctx.Args[1])
			return err
		}

		_, err = bot.DB.Exec(context.Background(), "update level_config set levels_enabled = $1 where id = $2", b, ctx.Guild.ID)
		if err != nil {
			return bot.Report(ctx, err)
		}
		_ = ctx.State.React(ctx.Message.ChannelID, ctx.Message.ID, "✅")

	case "dm_on_reward":
		b, err := strconv.ParseBool(ctx.Args[1])
		if err != nil {
			_, err = ctx.Replyc(bcr.ColourRed, "`%v` is not a valid setting for `dm_on_reward` (true or false)", ctx.Args[1])
			return err
		}

		_, err = bot.DB.Exec(context.Background(), "update level_config set dm_on_reward = $1 where id = $2", b, ctx.Guild.ID)
		if err != nil {
			return bot.Report(ctx, err)
		}
		_ = ctx.State.React(ctx.Message.ChannelID, ctx.Message.ID, "✅")

	case "between_xp":
		dur, err := time.ParseDuration(ctx.Args[1])
		if err != nil {
			_, err = ctx.Replyc(bcr.ColourRed, "`%v` is not a valid setting for `between_xp` (must be a duration)", ctx.Args[1])
			return err
		}

		_, err = bot.DB.Exec(context.Background(), "update level_config set between_xp = $1 where id = $2", dur, ctx.Guild.ID)
		if err != nil {
			return bot.Report(ctx, err)
		}
		_ = ctx.State.React(ctx.Message.ChannelID, ctx.Message.ID, "✅")

	case "reward_log", "nolevels_log":
		ch, err := ctx.ParseChannel(ctx.Args[1])
		if err != nil {
			_, err = ctx.Replyc(bcr.ColourRed, "`%v` is not a valid setting for `%v` (must be a text channel in this guild)", ctx.Args[1], strings.ToLower(ctx.Args[0]))
			return err
		}

		if ch.GuildID != ctx.Guild.ID || ch.Type != discord.GuildText {
			_, err = ctx.Replyc(bcr.ColourRed, "%v is not a valid channel for `%v` (must be a text channel in this guild)", ch.Mention(), strings.ToLower(ctx.Args[0]))
			return err
		}

		_, err = bot.DB.Exec(context.Background(), "update level_config set "+strings.ToLower(ctx.Args[0])+" = $1 where id = $2", ch.ID, ctx.Guild.ID)
		if err != nil {
			return bot.Report(ctx, err)
		}
		_ = ctx.State.React(ctx.Message.ChannelID, ctx.Message.ID, "✅")

	case "reward_text":
		text := strings.TrimSpace(strings.TrimPrefix(ctx.RawArgs, ctx.Args[0]))
		if text == ctx.RawArgs {
			text = strings.Join(ctx.Args[1:], " ")
		}
		if len(text) > 900 {
			_, err = ctx.Replyc(bcr.ColourRed, "Your text is too long! Maximum 900 characters.")
			return
		}

		_, err = bot.DB.Exec(context.Background(), "update level_config set reward_text = $1 where id = $2", text, ctx.Guild.ID)
		if err != nil {
			return bot.Report(ctx, err)
		}
		_ = ctx.State.React(ctx.Message.ChannelID, ctx.Message.ID, "✅")

	default:
		_, err = ctx.Replyc(bcr.ColourRed, "Sorry, but `%v` is not a valid configuration key. Valid keys are: enable, dm_on_reward, between_xp, nolevels_log, reward_log, reward_text.", ctx.Args[0])
		return
	}
	return
}

func (bot *Bot) blacklistAdd(ctx *bcr.Context) (err error) {
	// try parse channel
	ch, err := ctx.ParseChannel(ctx.RawArgs)
	if err == nil {
		if ch.GuildID != ctx.Guild.ID || (ch.Type != discord.GuildText && ch.Type != discord.GuildNews && ch.Type != discord.GuildCategory) {
			return ctx.SendX("Invalid channel provided, must be in this guild and be a text or category channel.")
		}

		if ch.Type == discord.GuildCategory {
			_, err := bot.DB.Exec(context.Background(), "update level_config set blocked_categories = array_append(blocked_categories, $1) where id = $2", ch.ID, ctx.Guild.ID)
			if err != nil {
				return bot.Report(ctx, err)
			}

			return ctx.SendfX("Added %v to the list of blocked categories!", ch.Name)
		} else {
			_, err := bot.DB.Exec(context.Background(), "update level_config set blocked_channels = array_append(blocked_channels, $1) where id = $2", ch.ID, ctx.Guild.ID)
			if err != nil {
				return bot.Report(ctx, err)
			}

			return ctx.SendfX("Added %v/#%v to the list of blocked channels!", ch.Mention(), ch.Name)
		}
	}

	// else try parsing role
	r, err := ctx.ParseRole(ctx.RawArgs)
	if err != nil {
		return ctx.SendX("Input is not a valid role or channel.")
	}

	_, err = bot.DB.Exec(context.Background(), "update level_config set blocked_roles = array_append(blocked_roles, $1) where id = $2", r.ID, ctx.Guild.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	return ctx.SendfX("Added @%v to the list of blocked roles!", r.Name)
}

func (bot *Bot) blacklistRemove(ctx *bcr.Context) (err error) {
	// try parse channel
	ch, err := ctx.ParseChannel(ctx.RawArgs)
	if err == nil {
		if ch.GuildID != ctx.Guild.ID || (ch.Type != discord.GuildText && ch.Type != discord.GuildNews && ch.Type != discord.GuildCategory) {
			return ctx.SendX("Invalid channel provided, must be in this guild and be a text or category channel.")
		}

		if ch.Type == discord.GuildCategory {
			_, err := bot.DB.Exec(context.Background(), "update level_config set blocked_categories = array_remove(blocked_categories, $1) where id = $2", ch.ID, ctx.Guild.ID)
			if err != nil {
				return bot.Report(ctx, err)
			}

			return ctx.SendfX("Remmoved %v from the list of blocked categories!", ch.Name)
		} else {
			_, err := bot.DB.Exec(context.Background(), "update level_config set blocked_channels = array_remove(blocked_channels, $1) where id = $2", ch.ID, ctx.Guild.ID)
			if err != nil {
				return bot.Report(ctx, err)
			}

			return ctx.SendfX("Removed %v/#%v from the list of blocked channels!", ch.Mention(), ch.Name)
		}
	}

	// else try parsing role
	r, err := ctx.ParseRole(ctx.RawArgs)
	if err != nil {
		return ctx.SendX("Input is not a valid role or channel.")
	}

	_, err = bot.DB.Exec(context.Background(), "update level_config set blocked_roles = array_remove(blocked_roles, $1) where id = $2", r.ID, ctx.Guild.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	return ctx.SendfX("Removed @%v from the list of blocked roles!", r.Name)
}

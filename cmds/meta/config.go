package meta

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) configList(ctx *bcr.Context) (err error) {
	if len(ctx.Args) > 0 {
		if strings.EqualFold(ctx.Args[0], "short") {
			var keys []string
			for k := range db.ConfigOptions {
				keys = append(keys, "`"+k+"`\n")
			}
			sort.Strings(keys)

			_, _, err = ctx.ButtonPages(bcr.StringPaginator("Configuration options", bot.Colour, keys, 20), 10*time.Minute)
			return
		}

		opt, ok := db.ConfigOptions[strings.ToLower(ctx.RawArgs)]
		if !ok {
			return ctx.SendfX("No setting named ``%v`` found.%v", bcr.EscapeBackticks(ctx.RawArgs), similarOptions(ctx.RawArgs))
		}

		return ctx.SendX("", bot.optionEmbed(
			bot.DB.Config.Get("prefix").ToString(),
			strings.ToLower(ctx.RawArgs),
			opt,
		))
	}

	var keys []string
	for k := range db.ConfigOptions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	prefix := bot.DB.Config.Get("prefix").ToString()

	var embeds []discord.Embed
	for i, key := range keys {
		e := bot.optionEmbed(prefix, key, db.ConfigOptions[key])
		e.Footer = &discord.EmbedFooter{
			Text: fmt.Sprintf("Setting %d of %d", i+1, len(keys)),
		}

		embeds = append(embeds, e)
	}

	_, _, err = ctx.ButtonPages(embeds, 20*time.Minute)
	return
}

func (bot *Bot) optionEmbed(prefix, name string, opt db.ConfigOption) discord.Embed {
	desc := strings.NewReplacer("{prefix}", prefix, "{docs}", db.ArikawaDocumentationBase).Replace(opt.Description)

	e := discord.Embed{
		Title:       "`" + name + "`",
		Description: desc,
		Fields: []discord.EmbedField{{
			Name:  "Type",
			Value: opt.Type.String(),
		}},
		Color: bot.Colour,
	}

	var def string
	switch opt.Type {
	case db.BoolOptionType, db.FloatOptionType, db.IntOptionType:
		def = fmt.Sprint(opt.DefaultValue)
	case db.StringOptionType:
		def = fmt.Sprintf("``%s``", opt.DefaultValue)
	case db.SnowflakeOptionType:
		return e
	}

	e.Fields = append(e.Fields, discord.EmbedField{
		Name:  "Default value",
		Value: def,
	})
	return e
}

func similarOptions(input string) string {
	if len(input) < 3 {
		return ""
	}

	s := "\nPerhaps you meant one of the following settings?\n> "
	var options []string

	for k := range db.ConfigOptions {
		if strings.Contains(k, input) {
			options = append(options, "`"+k+"`")
		}
	}

	if len(options) == 0 {
		return ""
	}
	sort.Strings(options)

	return s + strings.Join(options, ", ")
}

func (bot *Bot) configSet(ctx *bcr.Context) (err error) {
	var val interface{}

	name := strings.ToLower(ctx.Args[0])
	input := strings.TrimSpace(strings.TrimPrefix(ctx.RawArgs, ctx.Args[0]))
	if input == ctx.RawArgs {
		input = strings.Join(ctx.Args[1:], " ")
	}

	opt, ok := db.ConfigOptions[name]
	if !ok {
		return ctx.SendfX("Sorry, but ``%v`` is not a valid configuration setting.%v", bcr.EscapeBackticks(name), similarOptions(name))
	}

	switch opt.Type {
	case db.StringOptionType:
		val = input
	case db.BoolOptionType:
		switch strings.ToLower(input) {
		case "true", "t", "on", "yes", "1":
			val = true
		case "false", "f", "off", "no", "0":
			val = false
		default:
			return ctx.SendfX("Sorry, but ``%v`` is not a valid boolean setting.\nValid settings are `true` and `false`.", bcr.EscapeBackticks(input))
		}
	case db.IntOptionType:
		val, err = strconv.ParseInt(input, 10, 64)
		if err != nil {
			return ctx.SendfX("Sorry, but ``%v`` is not a valid integer.", bcr.EscapeBackticks(input))
		}
	case db.FloatOptionType:
		val, err = strconv.ParseFloat(input, 64)
		if err != nil {
			return ctx.SendfX("Sorry, but ``%v`` is not a valid floating number.", bcr.EscapeBackticks(input))
		}
	case db.SnowflakeOptionType:
		sf, mention, err := parseAnySlowflake(ctx, input)
		if err != nil {
			return ctx.SendfX("Sorry, but ``%v`` is not a valid snowflake.", bcr.EscapeBackticks(input))
		}
		input = mention
		val = sf
	default:
		return ctx.SendfX("This configuration setting is invalid! (expected type from 0-5, got type number %d)", int(opt.Type))
	}

	if len(opt.ValidValues) > 0 {
		isValid := false
		for _, v := range opt.ValidValues {
			if v == val {
				isValid = true
				break
			}
		}

		if !isValid {
			return ctx.SendfX("Sorry, but ``%v`` is not a valid option for ``%v``.\n(See the setting's description for valid options)", bcr.EscapeBackticks(input), name)
		}
	}

	err = bot.DB.Config.Set(name, val)
	if err != nil {
		return ctx.SendfX("Error setting setting: %v", err)
	}

	err = bot.DB.SyncConfig()
	if err != nil {
		common.Log.Errorf("Error syncing configuration with database: %v", err)
		return ctx.SendfX("Error syncing configuration with database: %v", err)
	}
	// fuck we need to set this or argument checks cause a panic
	bot.Router.Prefixes = []string{bot.DB.Config.Get("prefix").ToString()}

	_, err = ctx.Reply("Success! ``%v`` is now set to:\n>>> %v", bcr.EscapeBackticks(name), input)
	return
}

func (bot *Bot) configGet(ctx *bcr.Context) (err error) {
	name := strings.ToLower(ctx.Args[0])

	opt, ok := db.ConfigOptions[name]
	if !ok {
		return ctx.SendfX("Sorry, but ``%v`` is not a valid configuration setting.%v", bcr.EscapeBackticks(name), similarOptions(name))
	}
	current := bot.DB.Config.Get(name)

	_, exists := bot.DB.Config[name]

	e := discord.Embed{
		Title: "`" + name + "`",
		Color: bot.Colour,

		Fields: []discord.EmbedField{{
			Name:  "Default value?",
			Value: strconv.FormatBool(!exists),
		}},
	}

	switch opt.Type {
	case db.StringOptionType:
		if strings.Contains(current.ToString(), "\n") {
			e.Description = "```\n" + current.ToString() + "\n```"
		} else {
			e.Description = "``" + bcr.EscapeBackticks(current.ToString()) + "``"
		}
	case db.IntOptionType:
		e.Description = strconv.FormatInt(current.ToInt(), 10)
	case db.FloatOptionType:
		e.Description = strconv.FormatFloat(current.ToFloat(), 'f', -1, 64)
	case db.BoolOptionType:
		e.Description = strconv.FormatBool(current.ToBool())
	case db.SnowflakeOptionType:
		e.Description = formatAnySnowflake(ctx, current.ToSnowflake())
	default:
		e.Description = fmt.Sprintf("%v", current.ToInterface())
	}

	return ctx.SendX("", e)
}

func formatAnySnowflake(ctx *bcr.Context, sf discord.Snowflake) string {
	for _, r := range ctx.Guild.Roles {
		if r.ID == discord.RoleID(sf) {
			return r.Mention()
		}
	}

	chs, err := ctx.State.Channels(ctx.Message.GuildID)
	if err != nil {
		return discord.UserID(sf).Mention()
	}

	for _, ch := range chs {
		if ch.ID == discord.ChannelID(sf) {
			return ch.Mention()
		}
	}

	return discord.UserID(sf).Mention()
}

func parseAnySlowflake(ctx *bcr.Context, input string) (sf discord.Snowflake, mention string, err error) {
	// try channel
	ch, err := ctx.ParseChannel(input)
	if err == nil {
		return discord.Snowflake(ch.ID), ch.Mention(), nil
	}

	// try role
	r, err := ctx.ParseRole(input)
	if err == nil {
		return discord.Snowflake(r.ID), r.Mention(), nil
	}

	// try member
	m, err := ctx.ParseMember(input)
	if err == nil {
		return discord.Snowflake(m.User.ID), m.Mention(), nil
	}

	u, err := ctx.ParseUser(input)
	if err == nil {
		return discord.Snowflake(u.ID), u.Mention(), nil
	}

	i, err := strconv.ParseUint(input, 10, 64)
	if err == nil {
		return discord.Snowflake(i), input, nil
	}

	return 0, "", errInvalidSnowflake
}

const errInvalidSnowflake = errors.Sentinel("invalid snowflake")

package levels

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) addBackground(ctx *bcr.Context) (err error) {
	name := ctx.Args[0]
	emoji := strings.Split(strings.Trim(ctx.Args[1], "<>"), ":")
	desc := ctx.Args[2]

	var (
		emojiName string
		emojiID   *discord.Snowflake
	)

	switch len(emoji) {
	case 0:
		return ctx.SendX("Emoji is empty! This shouldn't happen.")
	case 1:
		emojiName = emoji[0]
	case 2:
		sf, _ := discord.ParseSnowflake(emoji[1])

		emojiName = emoji[0]
		emojiID = &sf
	default:
		sf, _ := discord.ParseSnowflake(emoji[2])

		emojiName = emoji[1]
		emojiID = &sf
	}

	if len(ctx.Message.Attachments) == 0 {
		return ctx.SendX("You must give an image to use as background!")
	}

	switch ctx.Message.Attachments[0].ContentType {
	case "image/jpeg", "image/png":
	default:
		return ctx.SendfX("You must give an *image* to use as a background!\nGot content type %q", ctx.Message.Attachments[0].ContentType)
	}

	resp, err := http.Get(ctx.Message.Attachments[0].URL)
	if err != nil {
		bot.SendError("Error downloading background: %v", err)
		return bot.Report(ctx, err)
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return bot.Report(ctx, err)
	}

	_, err = bot.DB.Exec(context.Background(), "insert into level_backgrounds (name, source, blob, emoji_name, emoji_id) values ($1, $2, $3, $4, $5)", name, desc, buf, emojiName, emojiID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	return ctx.SendfX("Added new banner **%v**!", name)
}

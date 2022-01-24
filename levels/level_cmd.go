package levels

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"strings"

	"github.com/AndreKR/multiface"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/disintegration/imaging"
	"github.com/dustin/go-humanize"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/common"
	"golang.org/x/image/font"
)

// yes this is taken from covebot. yes that means that this is slightly wrong. no i do not care, it took me like 30 minutes to figure this out the first time around, i'm not doing it again for like. two wrong pixels
var blankPixels = []int{96, 96, 96, 96, 85, 85, 85, 85, 74, 74, 74, 74, 68, 68, 68, 62, 62, 62, 62, 55, 55, 55, 55, 50, 50, 50, 50, 45, 45, 45, 45, 39, 39, 39, 39, 39, 39, 33, 33, 33, 33, 33, 33, 28, 28, 28, 28, 28, 24, 24, 24, 24, 24, 24, 20, 20, 20, 20, 20, 20, 16, 16, 16, 16, 16, 16, 16, 12, 12, 12, 12, 12, 12, 12, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 4, 4, 4, 4, 4, 4, 4, 4}

//go:embed fonts
var fontData embed.FS

var boldFont font.Face
var normalFont font.Face

func mustParse(path string) *truetype.Font {
	b, err := fontData.ReadFile(path)

	if err != nil {
		panic(err)
	}

	f, err := truetype.Parse(b)
	if err != nil {
		panic(err)
	}

	return f
}

func init() {
	// montserrat for most latin letters
	montserrat := mustParse("fonts/Montserrat-Medium.ttf")
	// noto as fallback for other characters
	noto := mustParse("fonts/NotoSans-Medium.ttf")
	// emoji fallback
	emoji := mustParse("fonts/NotoEmoji-Regular.ttf")

	mf := &multiface.Face{}

	// add montserrat font
	mf.AddTruetypeFace(truetype.NewFace(montserrat, &truetype.Options{
		Size: 60,
	}), montserrat)

	// add noto font
	mf.AddTruetypeFace(truetype.NewFace(noto, &truetype.Options{
		Size: 60,
	}), noto)

	// add noto emoji
	mf.AddTruetypeFace(truetype.NewFace(emoji, &truetype.Options{
		Size: 60,
	}), emoji)

	boldFont = mf

	normalFont = truetype.NewFace(
		mustParse("fonts/Montserrat-Regular.ttf"),
		&truetype.Options{
			Size: 40,
		},
	)
}

func (bot *Bot) levelCmd(ctx *bcr.Context) (err error) {
	u := &ctx.Author
	if len(ctx.Args) > 0 {
		u, err = ctx.ParseUser(strings.Join(ctx.Args, " "))
		if err != nil {
			_, err = ctx.Send("User not found.")
			return
		}
	}

	uc, err := bot.getUser(ctx.Message.GuildID, u.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	lvl := LevelFromXP(uc.XP)
	xpForNext := XPFromLevel(lvl + 1)
	xpForPrev := XPFromLevel(lvl)

	// get leaderboard (for rank)
	// filter the leaderboard to match the `leaderboard` command
	var rank int
	lb, err := bot.getLeaderboard(ctx.Message.GuildID, false)
	if err == nil {
		for i, uc := range lb {
			if uc.UserID == u.ID {
				rank = i + 1
				break
			}
		}
	}

	// get user colour + avatar URL
	clr := uc.Colour
	avatarURL := u.AvatarURLWithType(discord.PNGImage) + "?size=256"
	username := u.Username
	if clr == 0 && ctx.Guild != nil {
		m, err := ctx.State.Member(ctx.Guild.ID, u.ID)
		if err == nil {
			clr = discord.MemberColor(*ctx.Guild, *m)
			if m.Avatar != "" {
				avatarURL = m.AvatarURLWithType(discord.PNGImage, ctx.Message.GuildID) + "?size=256"
			}
			if m.Nick != "" {
				username = m.Nick
			}
		}
	}

	if useEmbed, _ := ctx.Flags.GetBool("embed"); useEmbed {
		e := bot.generateEmbed(
			ctx, username, avatarURL, clr,
			rank, lvl, uc.XP, xpForNext, xpForPrev,
		)
		return ctx.SendX("", e)
	}

	img, err := bot.generateImage(
		ctx, username, avatarURL, clr,
		rank, lvl, uc.XP, xpForNext, xpForPrev,
		bot.getBackground(uc.Background),
	)
	if err != nil {
		common.Log.Errorf("Error generating level card for %v, falling back to embed: %v", u.Tag(), err)

		e := bot.generateEmbed(
			ctx, username, avatarURL, clr,
			rank, lvl, uc.XP, xpForNext, xpForPrev,
		)
		return ctx.SendX("", e)
	}

	_, err = ctx.State.SendMessageComplex(ctx.Message.ChannelID, api.SendMessageData{
		Files: []sendpart.File{{
			Name:   "level_card.png",
			Reader: img,
		}},
	})
	return
}

const (
	width          = 1320
	height         = 400
	progressBarLen = width - 450
)

func (bot *Bot) generateImage(ctx *bcr.Context,
	name, avatarURL string, clr discord.Color,
	rank int, lvl, xp, xpForNext, xpForPrev int64,
	background []byte,
) (r io.Reader, err error) {

	img := gg.NewContext(width, height)

	// background
	if background != nil {
		bg, _, err := image.Decode(bytes.NewReader(background))
		if err == nil {
			bg = imaging.Resize(bg, width, 0, imaging.NearestNeighbor)
			img.DrawImageAnchored(bg, 0, 0, 0, 0)
		}
	}

	img.SetHexColor("#00000088")
	img.DrawRoundedRectangle(50, 50, width-100, height-100, 20)
	img.Fill()

	// fetch avatar
	resp, err := http.Get(avatarURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// decode avatar
	pfp, err := png.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	// use average of avatar if the user has no colour
	if clr == 0 {
		r, g, b, _ := AverageColour(pfp)

		clr = discord.Color(r)<<16 + discord.Color(g)<<8 + discord.Color(b)
	}

	// resize pfp to fit + crop to circle (shoddily)
	pfp = imaging.Resize(pfp, 256, 256, imaging.NearestNeighbor)

	pfpImg := gg.NewContextForImage(pfp)
	pfpImg.SetColor(color.RGBA{0, 0, 0, 0})

	for y := 0; y < len(blankPixels); y++ {
		for x := 0; x < blankPixels[y]; x++ {
			pfpImg.SetPixel(x, y)
			pfpImg.SetPixel(256-x, 256-y)
			pfpImg.SetPixel(x, 256-y)
			pfpImg.SetPixel(256-x, y)
		}
	}

	// draw pfp to context
	img.SetHexColor(fmt.Sprintf("#%06x", clr))
	img.DrawCircle(200, 200, 130)
	img.FillPreserve()

	img.DrawImageAnchored(pfpImg.Image(), 200, 200, 0.5, 0.5)

	img.SetLineWidth(5)
	img.Stroke()

	progress := xp - xpForPrev
	needed := xpForNext - xpForPrev

	p := float64(progress) / float64(needed)

	end := progressBarLen * p

	img.DrawRectangle(350, 275, end, 50)
	img.Fill()

	img.SetHexColor("#686868")
	img.DrawRectangle(350+end, 275, progressBarLen-end, 50)
	img.Fill()

	img.SetHexColor(fmt.Sprintf("#%06xCC", clr))

	img.SetColor(color.NRGBA{0xB5, 0xB5, 0xB5, 0xCC})

	img.DrawRectangle(350, 180, progressBarLen, 3)
	img.Fill()

	img.SetStrokeStyle(gg.NewSolidPattern(color.NRGBA{0xB5, 0xB5, 0xB5, 0xFF}))

	img.DrawRoundedRectangle(350, 275, progressBarLen, 50, 5)
	img.SetLineWidth(2)
	img.Stroke()

	img.SetHexColor("#ffffff")

	img.SetFontFace(boldFont)

	displayName := ""
	for i, r := range name {
		if i > 18 {
			displayName += "..."
			break
		}

		displayName += string(r)
	}

	// name
	img.DrawStringAnchored(displayName, 350, 120, 0, 0.5)

	// rank/xp
	img.SetFontFace(normalFont)

	if rank != 0 {
		img.DrawStringAnchored(fmt.Sprintf("Rank #%v", rank), width-100, 120, 1, 0.5)
	}

	img.DrawStringAnchored(fmt.Sprintf("Level %v", lvl), width-100, 200, 1, 1)

	img.DrawStringAnchored(fmt.Sprintf("%v%%", int64(p*100)), 350+(progressBarLen/2), 295, 0.5, 0.5)

	progressStr := fmt.Sprintf("%v/%v XP", humanize.Comma(progress), humanize.Comma(needed))

	img.DrawStringAnchored(progressStr, 350, 200, 0, 1)

	buf := new(bytes.Buffer)

	err = img.EncodePNG(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (bot *Bot) generateEmbed(ctx *bcr.Context,
	name, avatarURL string, clr discord.Color,
	rank int, lvl, xp, xpForNext, xpForPrev int64,
) discord.Embed {

	e := discord.Embed{
		Color:       clr,
		Title:       fmt.Sprintf("Level %v - Rank #%v", lvl, rank),
		Description: fmt.Sprintf("%v/%v XP", humanize.Comma(xp), humanize.Comma(XPFromLevel(lvl+1))),
		Thumbnail: &discord.EmbedThumbnail{
			URL: avatarURL,
		},
		Author: &discord.EmbedAuthor{
			Name: name,
		},
	}

	{
		progress := xp - xpForPrev
		needed := xpForNext - xpForPrev

		p := float64(progress) / float64(needed)

		percent := int64(p * 100)

		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "Progress to next level",
			Value: fmt.Sprintf("%v%% (%v/%v)", percent, humanize.Comma(progress), humanize.Comma(needed)),
		})
	}

	return e
}

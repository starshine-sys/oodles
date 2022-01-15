package meta

import (
	"bytes"

	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/starshine-sys/bcr"
	"gopkg.in/yaml.v3"
)

type exportTrack struct {
	Emoji       string   `yaml:"emoji"`
	Description string   `yaml:"description,flow"`
	Questions   []string `yaml:"questions,omitempty"`
}

func (bot *Bot) exportTracks(ctx *bcr.Context) (err error) {
	tracks, err := bot.DB.ApplicationTracks()
	if err != nil {
		return bot.Report(ctx, err)
	}

	export := make(map[string]exportTrack, len(tracks))

	for _, t := range tracks {
		qs, err := bot.DB.Questions(t.ID)
		if err != nil {
			return bot.Report(ctx, err)
		}

		var s []string
		for _, q := range qs {
			s = append(s, q.Question)
		}

		export[t.Name] = exportTrack{
			Emoji:       t.RawEmoji,
			Description: t.Description,
			Questions:   s,
		}
	}

	b, err := yaml.Marshal(export)
	if err != nil {
		return bot.Report(ctx, err)
	}

	return ctx.SendFiles("Here you go!", sendpart.File{
		Name:   "export.yaml",
		Reader: bytes.NewReader(b),
	})
}

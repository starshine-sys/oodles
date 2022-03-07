package fun

import (
	"fmt"
	"strings"

	"github.com/starshine-sys/bcr"
	"golang.org/x/text/unicode/runenames"
)

func (bot *Bot) charinfo(ctx *bcr.Context) error {
	var out string
	for _, r := range ctx.RawArgs {
		name := strings.Fields(runenames.Name(r))
		for i := range name {
			name[i] = strings.Title(strings.ToLower(name[i]))
		}

		out += fmt.Sprintf("`u%08X` %v\n", r, strings.Join(name, " "))
	}

	if len(out) > 2000 {
		return ctx.SendfX("Your input is too long (output was %v characters, max 2000)", len(out))
	}

	return ctx.SendX(out)
}

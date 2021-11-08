package common

import (
	"strings"
	"text/template"

	"github.com/diamondburned/arikawa/v3/discord"
)

// ExecTemplate executes a template string into a string.
func ExecTemplate(tmpl string, data interface{}) (string, error) {
	var b strings.Builder

	t, err := template.New("").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}

	err = t.Execute(&b, data)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

var funcMap template.FuncMap = template.FuncMap{
	"displayName": func(m *discord.Member) string {
		if m.Nick != "" {
			return m.Nick
		}
		return m.User.Username
	},
}

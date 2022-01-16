package meta

import (
	"fmt"
	"sort"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/pkgo/v2"
)

const roleFilter = "—"

func (bot *Bot) memberInfo(ctx *bcr.Context) (err error) {
	m := ctx.Member
	m.User = ctx.Author

	if len(ctx.Args) > 0 {
		m, err = ctx.ParseMember(ctx.RawArgs)
		if err != nil {
			return bot.userInfo(ctx)
		}
	}

	// filter the roles to only the ones the user has
	var rls bcr.Roles
	for _, gr := range ctx.Guild.Roles {
		for _, ur := range m.RoleIDs {
			if gr.ID == ur && !strings.Contains(gr.Name, roleFilter) {
				rls = append(rls, gr)
			}
		}
	}
	sort.Sort(rls)

	// get global info
	// can't do this with a single loop because the loop for colour has to break the moment it's found one
	var (
		userPerms   discord.Permissions
		highestRole = "No roles"
	)
	for _, r := range rls {
		userPerms |= r.Permissions
	}
	if len(rls) > 0 {
		highestRole = rls[0].Name
	}

	var perms []string
	if ctx.Guild.OwnerID == m.User.ID {
		perms = append(perms, "Server Owner")
		userPerms = userPerms.Add(discord.PermissionAll)
	}
	perms = append(perms, bcr.PermStringsFor(bcr.MajorPerms, userPerms)...)

	permString := strings.Join(perms, ", ")
	if len(permString) > 1000 {
		permString = permString[:1000] + "..."
	} else if permString == "" {
		permString = "No special permissions"
	}
	var b strings.Builder
	for i, r := range rls {
		if b.Len() > 900 {
			b.WriteString(fmt.Sprintf("\n```Too many roles to list (showing %v/%v)```", i, len(rls)))
			break
		}
		b.WriteString(r.Mention())
		if i != len(rls)-1 {
			b.WriteString(", ")
		}
	}
	if b.Len() == 0 {
		b.WriteString("No roles.")
	}

	colour := discord.MemberColor(*ctx.Guild, *m)
	if colour == 0 {
		colour = m.User.Accent
		if colour == 0 {
			colour = ctx.Router.EmbedColor
		}
	}

	avatarURL := m.User.AvatarURL()
	if m.Avatar != "" {
		avatarURL = m.AvatarURL(ctx.Guild.ID)
	}

	e := discord.Embed{
		Author: &discord.EmbedAuthor{
			Name: m.User.Username + "#" + m.User.Discriminator,
			Icon: m.User.AvatarURL(),
		},
		Thumbnail: &discord.EmbedThumbnail{
			URL: avatarURL,
		},
		Description: m.User.ID.String(),
		Color:       colour,

		Fields: []discord.EmbedField{
			{
				Name:  "User information for",
				Value: m.Mention(),
			},
			{
				Name:   "Avatar",
				Value:  fmt.Sprintf("[Link](%v?size=1024)", m.User.AvatarURL()),
				Inline: true,
			},
			{
				Name:   "Username",
				Value:  m.User.Tag(),
				Inline: true,
			},
			{
				Name: "Created at",
				Value: fmt.Sprintf("<t:%v:D> <t:%v:T>\n(%v)",
					m.User.ID.Time().Unix(), m.User.ID.Time().Unix(),
					bcr.HumanizeTime(bcr.DurationPrecisionMinutes, m.User.ID.Time().UTC()),
				),
			},
		},

		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("ID: %v", m.User.ID),
		},
		Timestamp: discord.NowTimestamp(),
	}

	if m.Avatar != "" {
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:   "Server avatar",
			Value:  fmt.Sprintf("[Link](%v?size=1024)", m.AvatarURL(ctx.Guild.ID)),
			Inline: true,
		})
	}

	e.Fields = append(e.Fields, []discord.EmbedField{
		{
			Name:   "Nickname",
			Value:  fmt.Sprintf("%v", If(m.Nick != "", m.Nick, m.User.Username)),
			Inline: true,
		},
		{
			Name:   "Highest role",
			Value:  highestRole,
			Inline: true,
		},
		{
			Name: "Joined at",
			Value: fmt.Sprintf("<t:%v:D> <t:%v:T>\n(%v)\n%v days after the server was created",
				m.Joined.Time().Unix(), m.Joined.Time().Unix(),
				bcr.HumanizeTime(bcr.DurationPrecisionMinutes, m.Joined.Time().UTC()),
				int(
					m.Joined.Time().Sub(ctx.Message.GuildID.Time()).Hours()/24,
				),
			),
		},
		{
			Name:  fmt.Sprintf("Roles (%v)", len(rls)),
			Value: b.String(),
		},
		{
			Name:  "Key permissions",
			Value: permString,
		},
	}...)

	if sys, err := bot.PK.Account(pkgo.Snowflake(m.User.ID)); err == nil {
		val := fmt.Sprintf("**ID:** %v", sys.ID)
		if sys.Name != "" {
			val += fmt.Sprintf("\n**Name:** %v", sys.Name)
		}
		if !sys.Created.IsZero() {
			val += fmt.Sprintf("\n**Created:** <t:%v:D> <t:%v:T>", sys.Created.Unix(), sys.Created.Unix())
		}

		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "PluralKit information",
			Value: val,
		})
	}

	_, err = ctx.Send("", e)
	return
}

func (bot *Bot) userInfo(ctx *bcr.Context) (err error) {
	u, err := ctx.ParseUser(ctx.RawArgs)
	if err != nil {
		return ctx.SendfX("User ``%v`` not found.", bcr.EscapeBackticks(ctx.RawArgs))
	}

	e := discord.Embed{
		Author: &discord.EmbedAuthor{
			Name: u.Username + "#" + u.Discriminator,
			Icon: u.AvatarURL(),
		},
		Thumbnail: &discord.EmbedThumbnail{
			URL: u.AvatarURL(),
		},
		Description: u.ID.String(),
		Color:       ctx.Router.EmbedColor,

		Fields: []discord.EmbedField{
			{
				Name:  "User information for",
				Value: u.Mention(),
			},
			{
				Name:   "Avatar",
				Value:  fmt.Sprintf("[Link](%v?size=1024)", u.AvatarURL()),
				Inline: true,
			},
			{
				Name:   "Username",
				Value:  u.Username + "#" + u.Discriminator,
				Inline: true,
			},
			{
				Name: "Created at",
				Value: fmt.Sprintf("<t:%v:D> <t:%v:T>\n(%v)",
					u.ID.Time().Unix(), u.ID.Time().Unix(),
					bcr.HumanizeTime(bcr.DurationPrecisionMinutes, u.ID.Time().UTC()),
				),
			},
		},

		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("ID: %v", u.ID),
		},
		Timestamp: discord.NowTimestamp(),
	}

	if u.Accent != 0 {
		e.Color = u.Accent
	}

	_, err = ctx.Send("", e)
	return
}

// If tries to emulate a ternary operation as well as possible
func If(b bool, t, f interface{}) interface{} {
	if b {
		return t
	}
	return f
}

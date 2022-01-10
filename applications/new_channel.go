package applications

import (
	"fmt"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/dustin/go-humanize/english"
	"github.com/mozillazg/go-unidecode"
)

const (
	errInvalidCategory    = errors.Sentinel("invalid category ID")
	errCategoryNotInGuild = errors.Sentinel("category not in bot guild")
	errNoManageRoles      = errors.Sentinel("bot needs manage roles in category")
)

const userPermissions = discord.PermissionViewChannel | discord.PermissionSendMessages | discord.PermissionReadMessageHistory | discord.PermissionAddReactions | discord.PermissionEmbedLinks | discord.PermissionAttachFiles

func (bot *Bot) newApplicationChannel(m discord.Member) (ch *discord.Channel, err error) {
	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	catID := bot.DB.Config.Get("application_category").ToChannelID()
	if !catID.IsValid() {
		return nil, errInvalidCategory
	}

	g, err := s.Guild(bot.DB.BotConfig.GuildID)
	if err != nil {
		return nil, err
	}

	cat, err := s.Channel(catID)
	if err != nil {
		return nil, err
	}

	if cat.GuildID != g.ID {
		return nil, errCategoryNotInGuild
	}

	botMember, err := s.Member(bot.DB.BotConfig.GuildID, bot.Router.Bot.ID)
	if err != nil {
		return nil, err
	}

	// perms
	if !hasManageRoles(*g, *botMember, *cat) {
		return nil, errNoManageRoles
	}

	return s.CreateChannel(bot.DB.BotConfig.GuildID, api.CreateChannelData{
		Name:       "✏️-app-" + unidecode.Unidecode(m.User.Username),
		Type:       discord.GuildText,
		Topic:      "Application channel for " + m.Mention(),
		CategoryID: cat.ID,
		Overwrites: append(cat.Overwrites, discord.Overwrite{
			ID:    discord.Snowflake(m.User.ID),
			Type:  discord.OverwriteMember,
			Allow: userPermissions,
		}),
		AuditLogReason: "Create application channel",
	})
}

func hasManageRoles(g discord.Guild, m discord.Member, ch discord.Channel) bool {
	// global perms--need admin
	for _, id := range m.RoleIDs {
		for _, r := range g.Roles {
			if r.ID == id && r.Permissions.Has(discord.PermissionAdministrator) {
				return true
			}
		}
	}

	// category perms--need explicit manage roles
	for _, o := range ch.Overwrites {
		// user overrides
		if o.ID == discord.Snowflake(m.User.ID) {
			if o.Allow.Has(discord.PermissionManageRoles) {
				return true
			}
		}

		if o.Type == discord.OverwriteMember {
			continue
		}

		// role overrides
		for _, r := range m.RoleIDs {
			if o.ID == discord.Snowflake(r) {
				if o.Allow.Has(discord.PermissionManageRoles) {
					return true
				}
			}
		}
	}

	return false
}

func (bot *Bot) sendInitialMessage(ch discord.ChannelID, m discord.Member) error {
	name := m.Nick
	if name == "" {
		name = m.User.Username
	}

	s, _ := bot.Router.StateFromGuildID(bot.DB.BotConfig.GuildID)

	g, err := s.Guild(bot.DB.BotConfig.GuildID)
	if err != nil {
		return err
	}

	tracks, err := bot.DB.ApplicationTracks()
	if err != nil {
		return err
	}

	tmpl := bot.DB.Config.Get("open_application_message").ToString()

	e := discord.Embed{
		Title: "Started application for " + name,
		Thumbnail: &discord.EmbedThumbnail{
			URL: bot.Router.Bot.AvatarURL() + "?size=512",
		},
		Color:       bot.Colour,
		Description: strings.ReplaceAll(tmpl, "{guild}", g.Name) + "\n\nAre you ",
	}

	var descs []string
	var buttons discord.ActionRowComponent

	for _, t := range tracks {
		descs = append(descs, fmt.Sprintf("%v (%s)", t.Description, t.Emoji()))
		buttons = append(buttons, &discord.ButtonComponent{
			Label:    t.Name,
			CustomID: discord.ComponentID("app-track:" + strconv.FormatInt(t.ID, 10)),
			Style:    discord.SecondaryButtonStyle(),
			Emoji: &discord.ComponentEmoji{
				Name:     t.Emoji().Name,
				ID:       t.Emoji().ID,
				Animated: t.Emoji().Animated,
			},
		})
	}

	e.Description += english.OxfordWordSeries(descs, "or") + "?"

	_, err = s.SendMessageComplex(ch, api.SendMessageData{
		Content: m.Mention(),
		Embeds:  []discord.Embed{e},
		Components: discord.ContainerComponents{
			&buttons,
		},
		AllowedMentions: &api.AllowedMentions{
			Users: []discord.UserID{m.User.ID},
		},
	})
	return err
}

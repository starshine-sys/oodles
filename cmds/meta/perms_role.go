package meta

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) permsList(ctx *bcr.Context) (err error) {
	e := discord.Embed{
		Color: bot.Colour,
		Title: "Permissions",
	}

	if len(bot.DB.Perms.Owner) > 0 {
		val := ""
		for _, o := range bot.DB.Perms.Owner {
			if o.Type == db.RolePermission {
				val += discord.RoleID(o.ID).Mention() + "\n"
			} else {
				val += discord.UserID(o.ID).Mention() + "\n"
			}
		}
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "`[5] OWNER`",
			Value: val,
		})
	}

	if len(bot.DB.Perms.Staff) > 0 {
		val := ""
		for _, o := range bot.DB.Perms.Staff {
			if o.Type == db.RolePermission {
				val += discord.RoleID(o.ID).Mention() + "\n"
			} else {
				val += discord.UserID(o.ID).Mention() + "\n"
			}
		}
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "`[4] STAFF`",
			Value: val,
		})
	}

	if len(bot.DB.Perms.Helper) > 0 {
		val := ""
		for _, o := range bot.DB.Perms.Helper {
			if o.Type == db.RolePermission {
				val += discord.RoleID(o.ID).Mention() + "\n"
			} else {
				val += discord.UserID(o.ID).Mention() + "\n"
			}
		}
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "`[3] HELPER`",
			Value: val,
		})
	}

	if len(bot.DB.Perms.User) > 0 {
		val := ""
		for _, o := range bot.DB.Perms.User {
			if o.Type == db.RolePermission {
				val += discord.RoleID(o.ID).Mention() + "\n"
			} else {
				val += discord.UserID(o.ID).Mention() + "\n"
			}
		}
		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "`[2] USER`",
			Value: val,
		})
	}

	if len(e.Fields) == 0 {
		e.Description = "No permission overrides configured."
	}

	return ctx.SendX("", e)
}

func (bot *Bot) permsAdd(ctx *bcr.Context) (err error) {
	var level db.PermissionLevel
	switch strings.ToLower(ctx.Args[0]) {
	case "user":
		level = db.UserLevel
	case "helper":
		level = db.HelperLevel
	case "staff":
		level = db.StaffLevel
	case "owner":
		level = db.OwnerLevel
	default:
		return ctx.SendfX("``%v`` is not a valid permission level.", bcr.EscapeBackticks(ctx.Args[0]))
	}

	var (
		id           discord.Snowflake
		overrideType db.PermissionType
		str          string
	)

	role, err := ctx.ParseRole(ctx.Args[1])
	if err == nil {
		if role.ID == discord.RoleID(ctx.Guild.ID) {
			return ctx.SendfX("You can't change the permissions of the `@everyone` role.")
		}

		id = discord.Snowflake(role.ID)
		overrideType = db.RolePermission
		str = role.Mention()
	} else {
		u, err := ctx.ParseUser(ctx.Args[1])
		if err == nil {
			id = discord.Snowflake(u.ID)
			overrideType = db.UserPermission
			str = u.Mention()
		} else {
			return ctx.SendfX("``%v`` is not a valid user or role.", bcr.EscapeBackticks(ctx.Args[1]))
		}
	}

	switch level {
	case db.UserLevel:
		for _, o := range bot.DB.Perms.User {
			if o.ID == id {
				return ctx.SendfX("%v already has `USER` permissions.", str)
			}
		}

		bot.DB.Perms.User = append(bot.DB.Perms.User, db.PermissionOverride{
			ID:   id,
			Type: overrideType,
		})
	case db.HelperLevel:
		for _, o := range bot.DB.Perms.Helper {
			if o.ID == id {
				return ctx.SendfX("%v already has `HELPER` permissions.", str)
			}
		}

		bot.DB.Perms.Helper = append(bot.DB.Perms.Helper, db.PermissionOverride{
			ID:   id,
			Type: overrideType,
		})
	case db.StaffLevel:
		for _, o := range bot.DB.Perms.Staff {
			if o.ID == id {
				return ctx.SendfX("%v already has `STAFF` permissions.", str)
			}
		}

		bot.DB.Perms.Staff = append(bot.DB.Perms.Staff, db.PermissionOverride{
			ID:   id,
			Type: overrideType,
		})
	case db.OwnerLevel:
		for _, o := range bot.DB.Perms.Owner {
			if o.ID == id {
				return ctx.SendfX("%v already has `OWNER` permissions.", str)
			}
		}

		bot.DB.Perms.Owner = append(bot.DB.Perms.Owner, db.PermissionOverride{
			ID:   id,
			Type: overrideType,
		})
	}

	err = bot.DB.SyncPerms()
	if err != nil {
		return bot.Report(ctx, err)
	}

	_, err = ctx.Reply("Added %s as a `%s`.", str, level)
	return err
}

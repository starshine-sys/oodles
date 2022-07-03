package applications

import (
	"context"
	"fmt"
	"sort"
	"time"

	"codeberg.org/eviedelta/detctime/durationparser"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/oodles/db"
)

// Time that users can be unverified + no app open before showing up in oo.unverified
const unverifiedTime = 4 * 24 * time.Hour

func (bot *Bot) unverified(ctx *bcr.Context) (err error) {
	dur := unverifiedTime
	if len(ctx.Args) > 0 {
		dur, err = durationparser.Parse(ctx.RawArgs)
		if err != nil {
			return ctx.SendfX("Couldn't parse ``%v`` as a valid duration.", bcr.EscapeBackticks(ctx.RawArgs))
		}
	}

	ms, err := ctx.State.Members(ctx.Message.GuildID)
	if err != nil {
		return ctx.SendfX("Couldn't fetch member list:\n> %v", err)
	}

	// filter members
	ms = reduceMembers(ms, func(m discord.Member) bool { return !m.User.Bot })
	if len(ms) == 0 {
		return ctx.SendfX("No non-bot members!")
	}

	ms = reduceMembers(ms, func(m discord.Member) bool { return m.Joined.Time().Before(time.Now().Add(-dur)) })
	if len(ms) == 0 {
		return ctx.SendfX("No members joined before %v ago at all!", bcr.HumanizeDuration(bcr.DurationPrecisionMinutes, dur))
	}

	verifiedRole := bot.DB.Config.Get("verified_role").ToRoleID()
	if verifiedRole.IsValid() {
		ms = reduceMembers(ms, func(m discord.Member) bool { return !containsRole(m.RoleIDs, verifiedRole) })
		if len(ms) == 0 {
			return ctx.SendfX("There are no unverified members who joined before %v ago!", bcr.HumanizeDuration(bcr.DurationPrecisionMinutes, dur))
		}
	}

	ms = reduceMembers(ms, func(m discord.Member) bool {
		for _, ov := range append(bot.DB.Perms.Staff, bot.DB.Perms.Helper...) {
			if ov.Type == db.UserPermission && ov.ID == discord.Snowflake(m.User.ID) {
				return false
			}

			if ov.Type == db.RolePermission && containsRole(m.RoleIDs, discord.RoleID(ov.ID)) {
				return false
			}
		}
		return true
	})

	ms = reduceMembers(ms, func(m discord.Member) bool {
		hasApp := false
		err = bot.DB.QueryRow(context.Background(), "select exists(select * from applications where user_id = $1 and closed = false)", m.User.ID).Scan(&hasApp)
		if err != nil {
			bot.SendError("Error checking application status for %v: %v", m.User.ID, err)
		}

		return !hasApp
	})

	if len(ms) == 0 {
		return ctx.SendfX("All members that joined before %v ago are verified or have an open application!", bcr.HumanizeDuration(bcr.DurationPrecisionMinutes, dur))
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].Joined.Time().Before(ms[j].Joined.Time()) })

	var s []string

	for _, m := range ms {
		s = append(s, fmt.Sprintf("%v/%v\nJoined <t:%v> (%v)\n\n", m.User.Tag(), m.Mention(), m.Joined.Time().Unix(), bcr.HumanizeTime(bcr.DurationPrecisionHours, m.Joined.Time())))
	}

	_, _, err = ctx.ButtonPages(
		bcr.StringPaginator(
			fmt.Sprintf("Unverified members without applications (%v)", len(ms)),
			bot.Colour, s, 10,
		), 15*time.Minute,
	)
	return err
}

// reduceMembers reduces the given slice of members, only returning members that match filter.
func reduceMembers(in []discord.Member, filter func(discord.Member) bool) (out []discord.Member) {
	for _, m := range in {
		if filter(m) {
			out = append(out, m)
		}
	}
	return out
}

func containsRole(ids []discord.RoleID, id discord.RoleID) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}

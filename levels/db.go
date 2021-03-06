package levels

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/starshine-sys/oodles/common"
)

// XPFromLevel returns the XP needed for the given level.
func XPFromLevel(level int64) (xp int64) {
	return int64(5.0 / 6.0 * float64(level) * (2*math.Pow(float64(level), 2) + 27*float64(level) + 91))
}

// LevelFromXP returns the level of the given XP
func LevelFromXP(xp int64) (level int64) {
	x := float64(xp + 1)
	pow := math.Cbrt(
		math.Sqrt(3)*math.Sqrt(3888.0*math.Pow(x, 2)+(291600.0*x)-207025.0) - 108.0*x - 4050.0,
	)

	res := (-pow/(2.0*math.Pow(3.0, 2.0/3.0)*math.Pow(5.0, 1.0/3.0)) -
		(61.0*math.Cbrt(5.0/3.0))/(2.0*pow) - (9.0 / 2.0))

	return int64(res)
}

type GuildConfig struct {
	ID discord.GuildID `json:"-"`

	BlockedChannels   []uint64 `json:"blocked_channels"`
	BlockedRoles      []uint64 `json:"blocked_roles"`
	BlockedCategories []uint64 `json:"blocked_categories"`

	RewardLog   discord.ChannelID `json:"reward_log"`
	NolevelsLog discord.ChannelID `json:"nolevels_log"`

	BetweenXP  time.Duration `json:"between_xp"`
	RewardText string        `json:"reward_text"`

	LevelsEnabled bool `json:"enabled"`
	DMOnReward    bool `json:"dm_on_reward"`
}

type LevelBackground struct {
	ID     int64
	Name   string
	Source string
	Blob   []byte

	EmojiName string
	EmojiID   *discord.EmojiID
}

type UserLevel struct {
	GuildID discord.GuildID `json:"-"`
	UserID  discord.UserID  `json:"user_id"`

	XP int64 `json:"xp"`

	Colour     discord.Color `json:"colour"`
	Background *int64        `json:"-"`

	LastXP time.Time `json:"-"`
}

type LevelReward struct {
	GuildID    discord.GuildID `json:"-"`
	Level      int64           `db:"lvl" json:"lvl"`
	RoleReward discord.RoleID  `json:"role"`
}

type Nolevels struct {
	GuildID discord.GuildID
	UserID  discord.UserID
	Expires bool
	Expiry  time.Time

	LogChannel discord.ChannelID // not in table, used for expiry loop
}

func (bot *Bot) getGuildConfig(guildID discord.GuildID) (gc GuildConfig, err error) {
	err = pgxscan.Get(context.Background(), bot.DB.Pool, &gc, "insert into level_config (id) values ($1) on conflict (id) do update set id = $1 returning *", guildID)
	return gc, err
}

func (bot *Bot) getUser(guildID discord.GuildID, userID discord.UserID) (l UserLevel, err error) {
	err = pgxscan.Get(context.Background(), bot.DB.Pool, &l, "insert into levels (guild_id, user_id) values ($1, $2) on conflict (guild_id, user_id) do update set guild_id = $1 returning *", guildID, userID)
	return l, err
}

func (bot *Bot) incrementXP(guildID discord.GuildID, userID discord.UserID) (newXP int64, err error) {
	xp := 15 + rand.Intn(12)

	err = bot.DB.Pool.QueryRow(context.Background(), "update levels set xp = xp + $4, last_xp = $3 where guild_id = $1 and user_id = $2 returning xp", guildID, userID, time.Now().UTC(), xp).Scan(&newXP)
	return
}

func (bot *Bot) getReward(guildID discord.GuildID, lvl int64) *LevelReward {
	r := LevelReward{}

	var exists bool
	_ = bot.DB.Pool.QueryRow(context.Background(), "select exists(select * from level_rewards where guild_id = $1 and lvl = $2)", guildID, lvl).Scan(&exists)
	if !exists {
		return nil
	}

	err := pgxscan.Get(context.Background(), bot.DB.Pool, &r, "select * from level_rewards where guild_id = $1 and lvl = $2", guildID, lvl)
	if err != nil {
		common.Log.Errorf("Error getting reward: %v", err)
		return nil
	}

	return &r
}

func (bot *Bot) getAllRewards(guildID discord.GuildID) (rwds []LevelReward, err error) {
	err = pgxscan.Select(context.Background(), bot.DB.Pool, &rwds, "select * from level_rewards where guild_id = $1 order by lvl asc", guildID)
	return
}

func (bot *Bot) getLeaderboard(guildID discord.GuildID, full bool) (lb []UserLevel, err error) {
	err = pgxscan.Select(context.Background(), bot.DB.Pool, &lb, "select * from levels where guild_id = $1 order by xp desc, user_id asc", guildID)
	if err != nil || full {
		return
	}

	s, _ := bot.Router.StateFromGuildID(guildID)

	ms, err := s.Members(guildID)
	if err != nil {
		return lb, nil
	}

	filtered := []UserLevel{}

	for _, l := range lb {
		for _, m := range ms {
			if m.User.ID == l.UserID {
				filtered = append(filtered, l)
				break
			}
		}
	}

	return filtered, nil
}

func (bot *Bot) isBlacklisted(guildID discord.GuildID, userID discord.UserID) (blacklisted bool) {
	err := bot.DB.Pool.QueryRow(context.Background(), "select exists(select user_id from nolevels where guild_id = $1 and user_id = $2)", guildID, userID).Scan(&blacklisted)
	if err != nil {
		common.Log.Errorf("Error checking if user is blacklisted from levels: %v", err)
	}

	return blacklisted
}

// getBackground gets the user's background, or a random background if they have none set, as []byte.
func (bot *Bot) getBackground(id *int64) (blob []byte) {
	if id != nil {
		err := bot.DB.QueryRow(context.Background(), "select blob from level_backgrounds where id = $1", id).Scan(&blob)
		if err != nil {
			common.Log.Errorf("Error getting background ID %v: %v", id, err)
			return nil
		}
		return blob
	}

	// get random background
	var lbs []LevelBackground
	err := pgxscan.Select(context.Background(), bot.DB, &lbs, "select * from level_backgrounds")
	if err != nil {
		common.Log.Errorf("Error getting backgrounds: %v", err)
		return nil
	}

	switch len(lbs) {
	case 0:
		return nil
	case 1:
		return lbs[0].Blob
	default:
		return lbs[rand.Intn(len(lbs))].Blob
	}
}

func (bot *Bot) backgroundMetadata() (lbs []LevelBackground, err error) {
	err = pgxscan.Select(context.Background(), bot.DB, &lbs, "select id, name, source, emoji_name, emoji_id from level_backgrounds")
	return
}

func (bot *Bot) bgExists(id int64) (exists bool) {
	err := bot.DB.QueryRow(context.Background(), "select exists(select id from level_backgrounds where id = $1)", id).Scan(&exists)
	if err != nil {
		common.Log.Errorf("Error getting background ID %v: %v", id, err)
		return false
	}
	return true
}

func (bot *Bot) background(id int64) (lb LevelBackground, err error) {
	err = pgxscan.Get(context.Background(), bot.DB, &lb, "select id, name, source, emoji_name, emoji_id from level_backgrounds where id = $1", id)
	return
}

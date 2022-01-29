package logging

import (
	"regexp"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
	"github.com/starshine-sys/pkgo/v2"
)

// these names are ignored if the webhook message has
// - m.Content == ""
// - len(m.Embeds) > 0
// - len(m.Attachments) == 0
var ignoreBotNames = [...]string{
	"Carl-bot Logging",
	"GitHub",
}

func (bot *Bot) messageCreate(m *gateway.MessageCreateEvent) {
	if m.GuildID != bot.DB.BotConfig.GuildID {
		return
	}

	for _, id := range pkBotsToCheck {
		if m.Author.ID == id {
			bot.pkMessageCreate(m)
		}
	}

	content := m.Content
	if m.Content == "" {
		content = "None"
	}

	msg := db.Message{
		ID:        m.ID,
		UserID:    m.Author.ID,
		ChannelID: m.ChannelID,
		ServerID:  m.GuildID,
		Username:  m.Author.Username + "#" + m.Author.Discriminator,

		Content: content,
	}

	err := bot.DB.InsertMessage(msg)
	if err != nil {
		common.Log.Errorf("Error inserting message: %v", err)
	}

	if !m.WebhookID.IsValid() {
		return
	}

	// filter out log messages [as best as we can]
	// - message with content or attachments is assumed to be a proxied message
	// - message must have embeds to be a log message
	if m.Content == "" && len(m.Embeds) > 0 && len(m.Attachments) == 0 {
		for _, name := range ignoreBotNames {
			if m.Author.Username == name {
				common.Log.Debugf("Ignoring webhook message by %v", m.Author.Tag())
			}
		}
	}

	// give some time for PK to process the message
	time.Sleep(2 * time.Second)

	// check if we handled this message already
	bot.HandledMessagesMu.Lock()
	if _, ok := bot.HandledMessages[m.ID]; ok {
		delete(bot.HandledMessages, m.ID)
		bot.HandledMessagesMu.Unlock()
		return
	}
	bot.HandledMessagesMu.Unlock()

	common.Log.Debugf("No PK info for webhook message %v, falling back to API", m.ID)

	pkm, err := bot.PK.Message(pkgo.Snowflake(m.ID))
	if err != nil {
		if v, ok := err.(*pkgo.PKAPIError); ok {
			if v.Code == pkgo.MessageNotFound {
				return
			}
		}

		common.Log.Errorf("Error getting message info from the PK API: %v", err)
	}

	bot.ProxiedTriggersMu.Lock()
	bot.ProxiedTriggers[discord.MessageID(pkm.Original)] = struct{}{}
	bot.ProxiedTriggersMu.Unlock()

	if pkm.System == nil || pkm.Member == nil {
		err = bot.DB.UpdateUserID(m.ID, discord.UserID(pkm.Sender))
	} else {
		err = bot.DB.UpdatePKInfo(m.ID, pkm.Sender, pkm.System.ID, pkm.Member.ID)
	}
	if err != nil {
		common.Log.Errorf("error updating pk info for %v: %v", m.ID, err)
	}
}

var pkBotsToCheck = []discord.UserID{466378653216014359}

var (
	linkRegex   = regexp.MustCompile(`^https:\/\/discord.com\/channels\/\d+\/(\d+)\/\d+$`)
	footerRegex = regexp.MustCompile(`^System ID: (\w{5,6}) \| Member ID: (\w{5,6}) \| Sender: .+ \((\d+)\) \| Message ID: (\d+) \| Original Message ID: (\d+)$`)
)

func (bot *Bot) pkMessageCreate(m *gateway.MessageCreateEvent) {
	// ensure we've actually stored the message
	time.Sleep(500 * time.Millisecond)

	// only handle events that are *probably* a log message
	if len(m.Embeds) == 0 || !linkRegex.MatchString(m.Content) {
		return
	}
	if m.Embeds[0].Footer == nil {
		return
	}
	if !footerRegex.MatchString(m.Embeds[0].Footer.Text) {
		return
	}

	groups := footerRegex.FindStringSubmatch(m.Embeds[0].Footer.Text)

	var (
		sysID    = groups[1]
		memberID = groups[2]
		userID   discord.UserID
		msgID    discord.MessageID
	)

	{
		sf, _ := discord.ParseSnowflake(groups[3])
		userID = discord.UserID(sf)
		sf, _ = discord.ParseSnowflake(groups[4])
		msgID = discord.MessageID(sf)

		originalMessageID, _ := discord.ParseSnowflake(groups[5])
		bot.ProxiedTriggersMu.Lock()
		bot.ProxiedTriggers[discord.MessageID(originalMessageID)] = struct{}{}
		bot.ProxiedTriggersMu.Unlock()
	}

	err := bot.DB.UpdatePKInfo(msgID, pkgo.Snowflake(userID), sysID, memberID)
	if err != nil {
		common.Log.Errorf("error updating pk info for %v: %v", msgID, err)
	}

	bot.HandledMessagesMu.Lock()
	bot.HandledMessages[msgID] = struct{}{}
	bot.HandledMessagesMu.Unlock()
}

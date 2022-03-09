package star

import (
	"context"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/jackc/pgx"
	"github.com/starshine-sys/oodles/common"
)

func (bot *Bot) reactionAdd(ev *gateway.MessageReactionAddEvent) {
	ch, err := bot.RootChannel(ev.ChannelID)
	if err != nil {
		common.Log.Errorf("getting root channel: %v", err)
		return
	}

	cfg, err := bot.override(ev.GuildID, ch.ID, ch.ParentID)
	if err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			return
		}

		common.Log.Errorf("getting starboard config: %v", err)
		return
	}

	if cfg.Disabled {
		common.Log.Debugf("channel or category for %v has disabled starboard", ev.ChannelID)
		return
	}

	if emojiString(ev.Emoji) != cfg.Emoji {
		common.Log.Debugf("emoji %v isn't starboard emoji (%v)", emojiString(ev.Emoji), cfg.Emoji)
		return
	}

	// get original message
	msg, err := bot.State.Message(ev.ChannelID, ev.MessageID)
	if err != nil {
		common.Log.Errorf("error getting message %v: %v", ev.MessageID, err)
		return
	}

	if msg.Author.ID == ev.UserID && !cfg.AllowSelfStar {
		common.Log.Debug("self starring isn't allowed")
		return
	}

	count, err := bot.addReaction(msg.ID, ev.UserID)
	if err != nil {
		common.Log.Error("storing starboard reaction:", err)
		return
	}

	err = bot.sendOrUpdateMessage(*msg, cfg, int(count))
	if err != nil {
		common.Log.Error("sending or updating starboard message:", err)
	}
}

func (bot *Bot) reactionRemoveAll(ev *gateway.MessageReactionRemoveAllEvent) {
	ch, err := bot.RootChannel(ev.ChannelID)
	if err != nil {
		common.Log.Errorf("getting root channel: %v", err)
		return
	}

	cfg, err := bot.override(ev.GuildID, ch.ID, ch.ParentID)
	if err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			return
		}

		common.Log.Errorf("getting starboard config: %v", err)
		return
	}

	if cfg.Disabled {
		common.Log.Debugf("channel or category for %v has disabled starboard", ev.ChannelID)
		return
	}

	_, err = bot.queries.RemoveAllReactions(context.Background(), int64(ev.MessageID))
	if err != nil {
		common.Log.Errorf("error removing all reactions: %v", err)
		return
	}

	sm, err := bot.queries.StarboardMessage(context.Background(), int64(ev.MessageID))
	if err != nil {
		common.Log.Error("getting starboard db entry:", err)
		return
	}

	err = bot.deleteMessage(cfg, sm.StarboardID)
	if err != nil {
		common.Log.Error("deleting starboard message:", err)
	}
}

func (bot *Bot) reactionRemoveEmoji(ev *gateway.MessageReactionRemoveEmojiEvent) {
	ch, err := bot.RootChannel(ev.ChannelID)
	if err != nil {
		common.Log.Errorf("getting root channel: %v", err)
		return
	}

	cfg, err := bot.override(ev.GuildID, ch.ID, ch.ParentID)
	if err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			return
		}

		common.Log.Errorf("getting starboard config: %v", err)
		return
	}

	if cfg.Disabled {
		common.Log.Debugf("channel or category for %v has disabled starboard", ev.ChannelID)
		return
	}

	if emojiString(ev.Emoji) != cfg.Emoji {
		common.Log.Debugf("emoji %v isn't starboard emoji (%v)", emojiString(ev.Emoji), cfg.Emoji)
		return
	}

	_, err = bot.queries.RemoveAllReactions(context.Background(), int64(ev.MessageID))
	if err != nil {
		common.Log.Errorf("error removing all reactions: %v", err)
		return
	}
}

func (bot *Bot) reactionRemove(ev *gateway.MessageReactionRemoveEvent) {
	ch, err := bot.RootChannel(ev.ChannelID)
	if err != nil {
		common.Log.Errorf("getting root channel: %v", err)
		return
	}

	cfg, err := bot.override(ev.GuildID, ch.ID, ch.ParentID)
	if err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			return
		}

		common.Log.Errorf("getting starboard config: %v", err)
		return
	}

	if cfg.Disabled {
		common.Log.Debugf("channel or category for %v has disabled starboard", ev.ChannelID)
		return
	}

	if emojiString(ev.Emoji) != cfg.Emoji {
		common.Log.Debugf("emoji %v isn't starboard emoji (%v)", emojiString(ev.Emoji), cfg.Emoji)
		return
	}

	// get original message
	msg, err := bot.State.Message(ev.ChannelID, ev.MessageID)
	if err != nil {
		common.Log.Errorf("error getting message %v: %v", ev.MessageID, err)
		return
	}

	if msg.Author.ID == ev.UserID && !cfg.AllowSelfStar {
		common.Log.Debug("self starring isn't allowed")
		return
	}

	count, err := bot.removeReaction(msg.ID, ev.UserID)
	if err != nil {
		common.Log.Error("storing starboard reaction:", err)
		return
	}

	err = bot.sendOrUpdateMessage(*msg, cfg, int(count))
	if err != nil {
		common.Log.Error("sending or updating starboard message:", err)
	}
}

func (bot *Bot) messageDelete(ev *gateway.MessageDeleteEvent) {
	_, err := bot.queries.RemoveStarboard(context.Background(), int64(ev.ID))
	if err != nil {
		common.Log.Error("deleting starboard message:", err)
	}
}

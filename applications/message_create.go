package applications

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/jackc/pgx/v4"
	"github.com/starshine-sys/oodles/common"
	"github.com/starshine-sys/oodles/db"
)

func (bot *Bot) messageCreate(m *gateway.MessageCreateEvent) {
	app, err := bot.DB.ChannelApplication(m.ChannelID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return
		}
		common.Log.Errorf("Couldn't fetch application: %v", err)
		return
	}

	// ignore bot messages
	if m.Author.Bot && m.Author.ID != bot.Router.Bot.ID {
		return
	}
	// save the message
	bot.saveMessage(app, m)

	// the rest only triggers for the application user
	if m.Author.ID != app.UserID {
		return
	}

	if app.Completed || app.TrackID == nil {
		return
	}

	qs, err := bot.DB.Questions(*app.TrackID)
	if err != nil {
		bot.SendError("Error fetching questions for app %v/track %v: %v", app.ID, *app.TrackID, err)
		return
	}

	if len(qs) <= app.Question {
		err = bot.completeApp(app, m)
		if err != nil {
			bot.SendError("Error completing app %v: %v", app.ID, err)
		}
		return
	}

	prev := qs[app.Question-1]
	if prev.LongAnswer && int64(len(strings.Fields(m.Content))) < bot.DB.Config.Get("long_answer_minimum").ToInt() {
		time.Sleep(3 * time.Second)

		tmpl := bot.DB.Config.Get("long_answer_message").ToString()
		msg := strings.ReplaceAll(tmpl, "{num}", strconv.FormatInt(bot.DB.Config.Get("long_answer_minimum").ToInt(), 10))

		err = bot.sendInterviewMessage(app, msg)
		if err != nil {
			bot.SendError("Error sending message in %v: %v", app.ChannelID.Mention(), err)
		}
		return
	}

	time.Sleep(3 * time.Second)
	err = bot.sendInterviewMessage(app, qs[app.Question].Question)
	if err != nil {
		bot.SendError("Error sending message in %v: %v", app.ChannelID.Mention(), err)
	}

	err = bot.DB.SetQuestionIndex(app.ID, app.Question+1)
	if err != nil {
		bot.SendError("Error incrementing question index for app %v: %v", app.ID, err)
	}
}

func (bot *Bot) saveMessage(app *db.Application, m *gateway.MessageCreateEvent) {
	resp := db.AppResponse{
		MessageID:     m.ID,
		UserID:        m.Author.ID,
		Username:      m.Author.Username,
		Discriminator: m.Author.Discriminator,
		Content:       m.Content,
		FromBot:       m.Author.Bot,
		FromStaff:     m.Author.ID != bot.Router.Bot.ID && m.Author.ID != app.UserID,
	}

	if len(m.Attachments) > 0 {
		resp.Content += "\n"
		for i, a := range m.Attachments {
			resp.Content += fmt.Sprintf("\n[Attachment #%d](%v)", i+1, a.URL)
		}
	}

	err := bot.DB.AddResponse(app.ID, resp)
	if err != nil {
		common.Log.Errorf("Error saving app message: %v", err)
	}
}

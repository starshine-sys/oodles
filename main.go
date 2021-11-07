package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/starshine-sys/oodles/bot"
	"github.com/starshine-sys/oodles/cmds/meta"
	"github.com/starshine-sys/oodles/common"
)

func main() {
	var conf common.BotConfig

	_, err := toml.DecodeFile("config.toml", &conf)
	if err != nil {
		common.Log.Fatalf("Error reading configuration file: %v", err)
	}

	common.Log.Infof("Starting Oodles version %v", common.Version)

	if !conf.LogChannel.IsValid() {
		common.Log.Warn("Warning: log_channel in config file is not valid. Errors will only be logged to console, and DMs will not be forwarded.")
	}

	b, err := bot.New(conf)
	if err != nil {
		common.Log.Fatalf("Error creating bot: %v", err)
	}

	// add commands/handlers
	meta.Init(b)

	state, _ := b.Router.StateFromGuildID(0)
	botUser, _ := state.Me()
	b.Router.Bot = botUser

	// open a connection to Discord
	if err = b.Start(context.Background()); err != nil {
		common.Log.Fatalf("Failed to connect: %v", err)
	}

	// Defer this to make sure that things are always cleanly shutdown even in the event of a crash
	defer func() {
		b.Router.ShardManager.Close()
		common.Log.Infof("Disconnected from Discord")
		b.DB.Close()
		common.Log.Infof("Database connection closed")
	}()

	common.Log.Info("Connected to Discord. Press Ctrl-C or send an interrupt signal to stop.")
	common.Log.Infof("User: %v (%v)", botUser.Tag(), botUser.ID)

	// alert in log if we don't receive a guild create event in time
	time.AfterFunc(5*time.Second, b.CheckIfReady)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	defer stop()

	select {
	case <-ctx.Done():
	}

	common.Log.Infof("Interrupt signal received. Shutting down...")
}

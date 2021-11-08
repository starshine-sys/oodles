package common

import (
	"os/exec"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is the global logger.
var Log = zap.S()

// Version is the git commit hash.
var Version = ""

func init() {
	// set up a logger
	zcfg := zap.NewProductionConfig()
	zcfg.Encoding = "console"
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zcfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	zcfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	zcfg.Level.SetLevel(zapcore.DebugLevel)

	logger, err := zcfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic(err)
	}

	Log = logger.Sugar()

	if Version == "" {
		Log.Info("Version not set, falling back to checking current directory.")

		git := exec.Command("git", "rev-parse", "--short", "HEAD")
		// ignoring errors *should* be fine? if there's no output we just fall back to "unknown"
		b, _ := git.Output()
		Version = strings.TrimSpace(string(b))
		if Version == "" {
			Version = "[unknown]"
		}
	}
}

// OpenApplication is the custom ID used for the "open application" button
const OpenApplication = "open_application"

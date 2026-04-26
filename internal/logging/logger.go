package logging

import (
	"log/slog"
	"os"
	"strings"
)

var levels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func ResolveLevel(value string) slog.Level {
	level, ok := levels[strings.ToLower(strings.TrimSpace(value))]
	if ok {
		return level
	}

	return slog.LevelInfo
}

func New(levelValue string) *slog.Logger {
	level := ResolveLevel(levelValue)

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

package logging

import (
	"log/slog"
	"os"
)

func New(level string) *slog.Logger {
	var minimum slog.Level
	if level == "debug" {
		minimum = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: minimum}))
}

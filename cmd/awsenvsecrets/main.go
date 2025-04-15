package main

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"syscall"

	awsenvsecrets "github.com/mackee/envsecrets/dist/aws"
)

func main() {
	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevel = "info"
	}
	ll := slog.LevelInfo
	switch logLevel {
	case "debug":
		ll = slog.LevelDebug
	case "warn":
		ll = slog.LevelWarn
	case "error":
		ll = slog.LevelError
	case "":
		ll = slog.LevelInfo
	default:
		slog.Warn("unknown log level, defaulting to info", slog.String("logLevel", logLevel))
	}
	slog.SetLogLoggerLevel(ll)
	ctx := context.Background()
	if err := awsenvsecrets.Load(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to load AWS secrets", slog.Any("error", err))
		os.Exit(1)
	}

	args := os.Args
	if len(args) == 0 {
		slog.ErrorContext(ctx, "no arguments provided")
		os.Exit(1)
	}

	cmd, err := exec.LookPath(args[1])
	if err != nil {
		slog.ErrorContext(ctx, "failed to find command", slog.Any("error", err))
		os.Exit(1)
	}
	if err := syscall.Exec(cmd, args[1:], os.Environ()); err != nil {
		slog.ErrorContext(ctx, "failed to exec command", slog.Any("error", err))
		os.Exit(1)
	}
}

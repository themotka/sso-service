package main

import (
	"github.com/themotka/sso-service/internal/app"
	"github.com/themotka/sso-service/internal/config"
	slogpretty "github.com/themotka/sso-service/pkg/logger/handler/slog-pretty"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	local = "local"
	dev   = "dev"
	prod  = "prod"
)

func main() {
	cfg := config.MustLoad()

	logger := initLogger(cfg.Environment)
	logger.Info("starting application", slog.Any("cfg", cfg))

	application := app.New(logger, cfg.Grpc.Port, cfg.StoragePath, cfg.TokenExpire)

	go application.Server.MustRun()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	stopSignal := <-shutdown

	logger.Info("shutdown signal received", slog.String("signal", stopSignal.String()))
	application.Server.Stop()
}

func initLogger(env string) *slog.Logger {
	var logger *slog.Logger
	switch env {
	case local:
		logger = initPrettyLogger(slog.LevelDebug, os.Stdout)
	case dev:
		logger = initPrettyLogger(slog.LevelDebug, os.Stdout)
	case prod:
		logger = initPrettyLogger(slog.LevelInfo, os.Stdout)
	}
	return logger
}

func initPrettyLogger(level slog.Level, writer io.Writer) *slog.Logger {
	operations := slogpretty.PrettyHandlerOptions{SlogOpts: &slog.HandlerOptions{
		Level: level,
	}}
	handler := operations.NewPrettyHandler(writer)
	return slog.New(handler)
}

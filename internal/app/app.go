package app

import (
	grpcapp "github.com/themotka/sso-service/internal/app/grpc"
	"log/slog"
	"time"
)

type App struct {
	Server *grpcapp.App
}

func New(logger *slog.Logger, port int, storage string, tokenTTL time.Duration) *App {
	//TODO: init storage + auth service
	grpcApp := grpcapp.New(logger, port)
	return &App{
		Server: grpcApp,
	}
}

package app

import (
	grpcapp "github.com/themotka/sso-service/internal/app/grpc"
	"github.com/themotka/sso-service/internal/services/oauth"
	"github.com/themotka/sso-service/internal/storage/sqlite"
	"log/slog"
	"time"
)

type App struct {
	Server *grpcapp.App
}

func New(logger *slog.Logger, port int, storage string, tokenTTL time.Duration) *App {
	database, err := sqlite.NewStorage(storage)
	if err != nil {
		panic(err)
	}
	authService := oauth.NewOAuth(logger, database, database, database, tokenTTL)
	grpcApp := grpcapp.New(logger, authService, port)
	return &App{
		Server: grpcApp,
	}
}

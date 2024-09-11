package grpcapp

import (
	"fmt"
	"github.com/themotka/sso-service/internal/gRPC/oauth"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	port       int
}

func New(logger *slog.Logger, authService oauth.OAuth, port int) *App {
	grpcServer := grpc.NewServer()
	oauth.RegisterServer(grpcServer, authService)
	return &App{
		log:        logger,
		grpcServer: grpcServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpc.run"
	logger := a.log.With(
		slog.String("operation", op),
		slog.Int("port", a.port))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	logger.Info("starting grpc server", slog.String("tcp:", listener.Addr().String()))

	err = a.grpcServer.Serve(listener)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.stop"
	a.log.With(
		slog.String("operation", op)).Info("stopping grpc server")
	//stop incoming requests, wait for the end of current
	a.grpcServer.GracefulStop()
}

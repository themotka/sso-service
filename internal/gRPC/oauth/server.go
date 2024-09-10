package oauth

import (
	"context"
	ssov1 "github.com/themotka/proto/gen/go/sso"
	"google.golang.org/grpc"
)

type server struct {
	ssov1.UnimplementedOAuthServer
	oauth OAuth
}

type OAuth interface {
	Login(ctx context.Context, email string, pass string) (*ssov1.LoginResponse, error)
}

func Register(gRPC *grpc.Server) {
	ssov1.RegisterOAuthServer(gRPC, &server{})
}

func (s *server) Login(ctx context.Context, request *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	return &ssov1.LoginResponse{JwtToken: request.GetPassword()}, nil
}

func (s *server) IsAdmin(ctx context.Context, request *ssov1.AdminRequest) (*ssov1.AdminResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) mustEmbedUnimplementedOAuthServer() {
	//TODO implement me
	panic("implement me")
}

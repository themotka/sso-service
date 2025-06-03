package oauth

import (
	"context"
	"errors"
	"github.com/asaskevich/govalidator"
	ssov1 "github.com/themotka/proto/gen/go/sso"
	"github.com/themotka/sso-service/internal/services/oauth"
	"github.com/themotka/sso-service/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	nullID = -1
)

type OAuth interface {
	Login(ctx context.Context, email string, pass string, appID int) (token string, err error)
	Register(ctx context.Context, email string, pass string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
}

type server struct {
	ssov1.UnimplementedOAuthServer
	oauth OAuth
}

func RegisterServer(gRPCServer *grpc.Server, oauth OAuth) {
	ssov1.RegisterOAuthServer(gRPCServer, &server{oauth: oauth})
}

func (s *server) Register(ctx context.Context, request *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(request); err != nil {
		return nil, err
	}
	userId, err := s.oauth.Register(ctx, request.GetEmail(), request.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &ssov1.RegisterResponse{UserId: userId}, nil
}

func (s *server) Login(ctx context.Context, request *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(request); err != nil {
		return nil, err
	}
	JWT, err := s.oauth.Login(ctx, request.GetEmail(), request.GetPassword(), int(request.GetAppId()))
	if err != nil {
		if errors.Is(err, oauth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &ssov1.LoginResponse{JwtToken: JWT}, nil
}

func (s *server) IsAdmin(ctx context.Context, request *ssov1.AdminRequest) (*ssov1.AdminResponse, error) {
	if err := validateIsAdmin(request); err != nil {
		return nil, err
	}
	isAdmin, err := s.oauth.IsAdmin(ctx, request.GetUserId())
	if err != nil {
		// TODO: error handling
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &ssov1.AdminResponse{IsAdmin: isAdmin}, nil
}

func (s *server) mustEmbedUnimplementedOAuthServer() {
	//TODO implement me
	panic("implement me")
}

func validateLogin(request *ssov1.LoginRequest) error {
	if ok := govalidator.IsEmail(request.GetEmail()); !ok {
		return status.Error(codes.InvalidArgument, "email is not valid")
	}
	if ok := govalidator.IsNull(request.GetPassword()); ok {
		return status.Error(codes.InvalidArgument, "password should not be empty")
	}
	if len(request.GetPassword()) < 8 {
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}
	if request.GetAppId() == nullID {
		return status.Error(codes.InvalidArgument, "app id should not be empty")
	}
	return nil
}

func validateRegister(request *ssov1.RegisterRequest) error {
	if ok := govalidator.IsEmail(request.GetEmail()); !ok {
		return status.Error(codes.InvalidArgument, "email is not valid")
	}
	if ok := govalidator.IsNull(request.GetPassword()); ok {
		return status.Error(codes.InvalidArgument, "password should not be empty")
	}
	if len(request.GetPassword()) < 8 {
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}
	return nil
}

func validateIsAdmin(request *ssov1.AdminRequest) error {
	if request.GetUserId() != nullID {
		return status.Error(codes.InvalidArgument, "userId should not be empty")
	}
	return nil
}

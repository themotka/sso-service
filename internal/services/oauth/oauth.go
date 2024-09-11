package oauth

import (
	"context"
	"errors"
	"fmt"
	"github.com/themotka/sso-service/internal/domain/model"
	"github.com/themotka/sso-service/internal/lib/jwt"
	"github.com/themotka/sso-service/internal/storage"
	"github.com/themotka/sso-service/pkg/logger/slogg"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppId       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
)

type OAuth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenExpire  time.Duration
}

func NewOAuth(log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenExpire time.Duration,
) *OAuth {
	return &OAuth{log: log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenExpire:  tokenExpire,
	}
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (user model.User, err error)
	IsAdmin(ctx context.Context, uid int64) (isAdmin bool, err error)
}

type AppProvider interface {
	App(ctx context.Context, appId int) (user model.App, err error)
}

func (o *OAuth) Login(ctx context.Context, email string, pass string, appID int,
) (token string, err error) {
	const op = "services.oauth.Login"
	log := o.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("logging in user")

	usr, err := o.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			o.log.Warn("user not found", slogg.Err(err))
		} else if errors.Is(err, storage.ErrUserAlreadyExists) {
			o.log.Warn("user already exists", slogg.Err(err))
		} else if errors.Is(err, storage.ErrAppNotFound) {
			o.log.Warn("app not found", slogg.Err(err))
		} else {
			o.log.Error("error getting user", slogg.Err(err))
		}
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	err = bcrypt.CompareHashAndPassword(usr.PassHash, []byte(pass))
	if err != nil {
		o.log.Info("invalid credentials", slogg.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := o.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err = jwt.GenerateToken(usr, app, o.tokenExpire)
	if err != nil {
		o.log.Error("error generating token", slogg.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}

func (o *OAuth) Register(ctx context.Context, email string, pass string,
) (userID int64, err error) {
	const op = "services.oauth.Register"
	log := o.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("registering user")

	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("error generating bcrypt hash", slogg.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := o.userSaver.SaveUser(ctx, email, bcryptHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Warn("user already exists", slogg.Err(err))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("error saving user", slogg.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user created")

	return id, nil
}

func (o *OAuth) IsAdmin(ctx context.Context, userID int64,
) (isAdmin bool, err error) {
	const op = "services.oauth.IsAdmin"
	log := o.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	log.Info("checking user permissions")

	isAdmin, err = o.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", slogg.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppId)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked, isAdmin=", slog.Bool("isAdmin", isAdmin))

	return isAdmin, nil
}

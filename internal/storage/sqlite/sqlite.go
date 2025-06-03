package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/themotka/sso-service/internal/domain/model"
	"github.com/themotka/sso-service/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.NewStorage"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) User(ctx context.Context, email string) (user model.User, err error) {
	const op = "storage.sqlite.User"

	prepare, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email=?")
	if err != nil {
		return model.User{}, fmt.Errorf("%s: %w", op, err)
	}
	user = model.User{}
	err = prepare.QueryRowContext(ctx, email).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return model.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, uid int64) (isAdmin bool, err error) {
	const op = "storage.sqlite.IsAdmin"

	prepare, err := s.db.Prepare("SELECT flag FROM admin_info WHERE id=?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	err = prepare.QueryRowContext(ctx, uid).Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error) {
	const op = "storage.sqlite.SaveUser"

	prepare, err := s.db.Prepare("INSERT INTO users (email, pass_hash) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	exec, err := prepare.ExecContext(ctx, email, passHash)
	if err != nil {
		var sqlErr *sqlite3.Error

		if errors.As(err, &sqlErr) && errors.Is(sqlErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserAlreadyExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := exec.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *Storage) App(ctx context.Context, id int) (model.App, error) {
	const op = "storage.sqlite.App"

	prepare, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
	if err != nil {
		return model.App{}, fmt.Errorf("%s: %w", op, err)
	}

	var app model.App
	err = prepare.QueryRowContext(ctx, id).Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return model.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

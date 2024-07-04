package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/zaytcevcom/msa/internal/storage"
)

type App struct {
	logger  Logger
	storage Storage
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type Storage interface {
	GetByID(ctx context.Context, id int) *storage.User
	Create(ctx context.Context, user storage.UserCreateDTO) (int, error)
	Update(ctx context.Context, id int, user storage.User) error
	Delete(ctx context.Context, id int) error
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) Health(_ context.Context) interface{} {
	return struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}
}

func (a *App) GetByID(ctx context.Context, id int) (*storage.User, error) {
	user := a.storage.GetByID(ctx, id)

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (a *App) Create(
	ctx context.Context,
	username string,
	password string,
	firstName string,
	lastName string,
	email string,
	phone string,
) (int, error) {
	hash := sha256.Sum256([]byte(password))
	hashString := hex.EncodeToString(hash[:])

	user := storage.UserCreateDTO{
		Username:     username,
		PasswordHash: hashString,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		Phone:        phone,
	}

	id, err := a.storage.Create(ctx, user)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (a *App) Update(ctx context.Context, id int, user storage.User) error {
	return a.storage.Update(ctx, id, user)
}

func (a *App) Delete(ctx context.Context, id int) error {
	return a.storage.Delete(ctx, id)
}

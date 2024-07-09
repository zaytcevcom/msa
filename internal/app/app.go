package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zaytcevcom/msa/internal/storage/user"
)

type App struct {
	logger  Logger
	storage Storage
	broker  MessageBroker
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type Storage interface {
	GetByID(ctx context.Context, id int) *storageuser.Entity
	Create(ctx context.Context, user storageuser.CreateDTO) (int, error)
	Update(ctx context.Context, id int, user storageuser.Entity) error
	Delete(ctx context.Context, id int) error
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
}

type MessageBroker interface {
	Publish(body string) error
	Close() error
}

func New(logger Logger, storage Storage, broker MessageBroker) *App {
	return &App{
		logger:  logger,
		storage: storage,
		broker:  broker,
	}
}

func (a *App) Health(_ context.Context) interface{} {
	return struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}
}

func (a *App) GetByID(ctx context.Context, id int) (*storageuser.Entity, error) {
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

	user := storageuser.CreateDTO{
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

	event := storageuser.UserEvent{
		Type:   storageuser.UserCreated,
		UserID: id,
	}

	msg, err := json.Marshal(event)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Failed marshal: %s", err))
		return id, nil
	}

	err = a.broker.Publish(string(msg))
	if err != nil {
		a.logger.Error(fmt.Sprintf("Failed publish: %s", err))
		return id, nil
	}

	return id, nil
}

func (a *App) Update(ctx context.Context, id int, user storageuser.Entity) error {
	return a.storage.Update(ctx, id, user)
}

func (a *App) Delete(ctx context.Context, id int) error {
	return a.storage.Delete(ctx, id)
}

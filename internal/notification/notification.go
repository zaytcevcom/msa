package notification

import (
	"context"

	storagenotification "github.com/zaytcevcom/msa/internal/storage/notification"
)

type Notification struct {
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
	GetByUserID(ctx context.Context, userID int) []storagenotification.EntityNotification
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
}

func New(logger Logger, storage Storage) *Notification {
	return &Notification{
		logger:  logger,
		storage: storage,
	}
}

func (n *Notification) Health(_ context.Context) interface{} {
	return struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}
}

func (n *Notification) GetByUserID(ctx context.Context, userID int) ([]storagenotification.EntityNotification, error) {
	return n.storage.GetByUserID(ctx, userID), nil
}

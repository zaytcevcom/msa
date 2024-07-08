package notificationsender

import (
	"context"
	"encoding/json"
)

type Sender struct {
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
	Create(ctx context.Context, userID int, email string, text string) error
}

type MessageBroker interface {
	Subscribe(handler func(body []byte) error) error
	Close() error
}

type Data struct {
	Type   string `json:"type,omitempty"`
	UserID int    `json:"userId,omitempty"`
	Email  string `json:"email,omitempty"`
	Text   string `json:"text,omitempty"`
}

func New(logger Logger, storage Storage, broker MessageBroker) *Sender {
	return &Sender{
		logger:  logger,
		storage: storage,
		broker:  broker,
	}
}

func (s Sender) Start() error {
	s.logger.Debug("Notification sender started!")

	return s.broker.Subscribe(func(body []byte) error {
		s.logger.Info(string(body))

		var data Data

		err := json.Unmarshal(body, &data)
		if err != nil {
			s.logger.Error(err.Error())
			return nil
		}

		err = s.storage.Create(context.Background(), data.UserID, data.Email, data.Text)
		if err != nil {
			s.logger.Error(err.Error())
			return nil
		}

		return nil
	})
}

func (s Sender) Stop() error {
	return s.broker.Close()
}

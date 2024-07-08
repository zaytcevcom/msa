package accountcreator

import (
	"context"
	"encoding/json"
)

type Creator struct {
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
	Create(ctx context.Context, userID int) (int, error)
}

type MessageBroker interface {
	Subscribe(handler func(body []byte) error) error
	Close() error
}

type Data struct {
	Type   string `json:"type,omitempty"`
	UserID int    `json:"userId,omitempty"`
}

func New(logger Logger, storage Storage, broker MessageBroker) *Creator {
	return &Creator{
		logger:  logger,
		storage: storage,
		broker:  broker,
	}
}

func (c *Creator) Start() error {
	c.logger.Debug("Account creator started!")

	return c.broker.Subscribe(func(body []byte) error {
		c.logger.Info(string(body))

		var data Data

		err := json.Unmarshal(body, &data)
		if err != nil {
			c.logger.Error(err.Error())
			return nil
		}

		_, err = c.storage.Create(context.Background(), data.UserID)
		if err != nil {
			c.logger.Error(err.Error())
			return nil
		}

		return nil
	})
}

func (c *Creator) Stop() error {
	return c.broker.Close()
}

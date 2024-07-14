package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zaytcevcom/msa/internal/order"
)

type OrderConsumer struct {
	logger   Logger
	storage  Storage
	consumer BrokerConsumer
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type Storage interface {
	ChangeStatus(ctx context.Context, orderID int, status order.Status) error
}

type BrokerConsumer interface {
	Subscribe(handler func(body []byte) error) error
	Close() error
}

type Data struct {
	OrderID int       `json:"orderId,omitempty"`
	Time    time.Time `json:"time,omitempty"`
}

func New(
	logger Logger,
	storage Storage,
	consumer BrokerConsumer,
) *OrderConsumer {
	return &OrderConsumer{
		logger:   logger,
		storage:  storage,
		consumer: consumer,
	}
}

func (c *OrderConsumer) Start() error {
	c.logger.Debug("Order consumer started!")

	return c.consumer.Subscribe(func(body []byte) error {
		c.logger.Info(string(body))

		var data Data

		err := json.Unmarshal(body, &data)
		if err != nil {
			c.logger.Error(err.Error())
			return err
		}

		err = c.storage.ChangeStatus(context.Background(), data.OrderID, order.NeedDelivery)
		if err != nil {
			c.logger.Error(err.Error())
			return err
		}

		return nil
	})
}

func (c *OrderConsumer) Stop() {
	_ = c.consumer.Close()
}

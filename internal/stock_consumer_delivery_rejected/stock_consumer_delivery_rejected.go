package consumer

import (
	"context"
	"encoding/json"
	"fmt"
)

type Consumer struct {
	logger   Logger
	storage  Storage
	consumer BrokerConsumer
	producer BrokerProducer
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type Storage interface {
	CancelReserve(ctx context.Context, orderID int) error
}

type BrokerConsumer interface {
	Subscribe(handler func(body []byte) error) error
	Close() error
}

type BrokerProducer interface {
	Publish(body string) error
	Close() error
}

type Data struct {
	OrderID int `json:"orderId,omitempty"`
}

func New(
	logger Logger,
	storage Storage,
	consumer BrokerConsumer,
	producer BrokerProducer,
) *Consumer {
	return &Consumer{
		logger:   logger,
		storage:  storage,
		consumer: consumer,
		producer: producer,
	}
}

func (c *Consumer) Start() error {
	c.logger.Debug("Stock consumer started!")

	return c.consumer.Subscribe(func(body []byte) error {
		c.logger.Info(string(body))

		var data Data

		err := json.Unmarshal(body, &data)
		if err != nil {
			c.logger.Error(err.Error())
			return err
		}

		err = c.storage.CancelReserve(context.Background(), data.OrderID)
		if err != nil {
			c.logger.Error(err.Error())
			return err
		}

		sendNotification(data, c.logger, c.producer)

		return nil
	})
}

func (c *Consumer) Stop() {
	_ = c.consumer.Close()
	_ = c.producer.Close()
}

func sendNotification(event interface{}, logger Logger, broker BrokerProducer) {
	msg, err := json.Marshal(event)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed marshal: %s", err))
		return
	}

	err = broker.Publish(string(msg))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed publish: %s", err))
		return
	}
}

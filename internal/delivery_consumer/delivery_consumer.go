package deliveryconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	storagedelivery "github.com/zaytcevcom/msa/internal/storage/delivery"
)

type DeliveryConsumer struct {
	logger          Logger
	storage         Storage
	consumer        BrokerConsumer
	producerSuccess BrokerProducer
	producerReject  BrokerProducer
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type Storage interface {
	IsTimeAvailable(_ context.Context, time time.Time) (bool, error)
	Reserve(ctx context.Context, orderID int, time time.Time) (int, error)
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
	producerSuccess BrokerProducer,
	producerReject BrokerProducer,
) *DeliveryConsumer {
	return &DeliveryConsumer{
		logger:          logger,
		storage:         storage,
		consumer:        consumer,
		producerSuccess: producerSuccess,
		producerReject:  producerReject,
	}
}

func (c *DeliveryConsumer) Start() error {
	c.logger.Debug("Delivery consumer started!")

	return c.consumer.Subscribe(func(body []byte) error {
		c.logger.Info(string(body))

		var data Data

		err := json.Unmarshal(body, &data)
		if err != nil {
			c.logger.Error("[1] " + err.Error())
			return err
		}

		eventNotReserved := storagedelivery.DeliveryNotReservedEvent{
			OrderID: data.OrderID,
		}

		// todo: hardcoded
		timeDelivery := time.Unix(1721314800, 0)

		isAvailable, err := c.storage.IsTimeAvailable(context.Background(), timeDelivery)
		if err != nil {
			sendNotification(eventNotReserved, c.logger, c.producerReject)
			c.logger.Error("[2] " + err.Error())
			return err
		}

		if !isAvailable {
			sendNotification(eventNotReserved, c.logger, c.producerReject)
			c.logger.Error("[3] " + err.Error())
			return nil
		}

		id, err := c.storage.Reserve(context.Background(), data.OrderID, timeDelivery)
		if err != nil {
			sendNotification(eventNotReserved, c.logger, c.producerReject)
			c.logger.Error("[4] " + err.Error())
			return err
		}

		eventReserved := storagedelivery.DeliveryReservedEvent{
			OrderID: id,
		}
		sendNotification(eventReserved, c.logger, c.producerSuccess)

		return nil
	})
}

func (c *DeliveryConsumer) Stop() {
	_ = c.consumer.Close()
	_ = c.producerReject.Close()
	_ = c.producerSuccess.Close()
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

package stockconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	storagestock "github.com/zaytcevcom/msa/internal/storage/stock"
)

type StockConsumer struct {
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
	GetCountAvailable(_ context.Context, productID int) (int, error)
	Reserve(ctx context.Context, orderID int, productID int, count int) (int, error)
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
	OrderID   int `json:"orderId,omitempty"`
	ProductID int `json:"productId,omitempty"`
	Count     int `json:"count,omitempty"`
}

func New(
	logger Logger,
	storage Storage,
	consumer BrokerConsumer,
	producerSuccess BrokerProducer,
	producerReject BrokerProducer,
) *StockConsumer {
	return &StockConsumer{
		logger:          logger,
		storage:         storage,
		consumer:        consumer,
		producerSuccess: producerSuccess,
		producerReject:  producerReject,
	}
}

func (c *StockConsumer) Start() error {
	c.logger.Debug("Stock consumer started!")

	return c.consumer.Subscribe(func(body []byte) error {
		c.logger.Info(string(body))

		var data Data

		err := json.Unmarshal(body, &data)
		if err != nil {
			c.logger.Error(err.Error())
			return err
		}

		eventNotReserved := storagestock.StockNotReservedEvent{
			OrderID: data.OrderID,
		}

		available, err := c.storage.GetCountAvailable(context.Background(), data.ProductID)
		if err != nil {
			sendNotification(eventNotReserved, c.logger, c.producerReject)
			c.logger.Error(err.Error())
			return err
		}

		// todo: added products
		if available < data.Count {
			c.logger.Warn("Product count: " + strconv.Itoa(available))
			//	sendNotification(eventNotReserved, c.logger, c.producerReject)
			//	return nil
		}

		id, err := c.storage.Reserve(context.Background(), data.OrderID, data.ProductID, data.Count)
		if err != nil {
			sendNotification(eventNotReserved, c.logger, c.producerReject)
			c.logger.Error(err.Error())
			return err
		}

		eventReserved := storagestock.StockReservedEvent{
			OrderID: id,
		}
		sendNotification(eventReserved, c.logger, c.producerSuccess)

		return nil
	})
}

func (c *StockConsumer) Stop() {
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

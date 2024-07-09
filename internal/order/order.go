package order

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	ordermiddleware "github.com/zaytcevcom/msa/internal/server/order/middleware"
	storageorder "github.com/zaytcevcom/msa/internal/storage/order"
)

type Order struct {
	logger  Logger
	storage Storage
	broker  MessageBroker
	cache   Cache
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type Storage interface {
	Create(ctx context.Context, userID int, productID int, sum float64, status Status, time time.Time) (int, error)
	ChangeStatus(ctx context.Context, orderID int, status Status) error
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
}

type MessageBroker interface {
	Publish(body string) error
	Close() error
}

type Cache interface {
	Get(key string) (string, bool)
	Set(key string, value string) bool
}

type Status int

const (
	Pending Status = iota
	Wait
	Done
)

func New(logger Logger, storage Storage, broker MessageBroker, cache Cache) *Order {
	return &Order{
		logger:  logger,
		storage: storage,
		broker:  broker,
		cache:   cache,
	}
}

func (o *Order) Health(_ context.Context) interface{} {
	return struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}
}

func (o *Order) Create(ctx context.Context, userID int, productID int, sum float64, email string) (int, error) {
	requestID := getRequestID(ctx)

	if requestID != "" {
		val, found := o.cache.Get(requestID)
		if found {
			id, _ := strconv.Atoi(val)
			return id, nil
		}
	}

	id, err := o.storage.Create(ctx, userID, productID, sum, Pending, time.Now())
	if err != nil {
		return 0, err
	}

	message := "Need pay order!"
	isPayed, err := pay(ctx, id, sum)
	if err != nil {
		o.logger.Info(err.Error())
	}

	if isPayed {
		err = o.storage.ChangeStatus(ctx, id, Wait)
		if err != nil {
			return 0, err
		}

		message = "Order success paid!"
	}

	event := storageorder.OrderEvent{
		Type:    storageorder.OrderCreated,
		OrderID: id,
		UserID:  userID,
		Email:   email,
		Text:    message,
	}
	sendNotification(event, o.logger, o.broker)

	if requestID != "" {
		o.cache.Set(requestID, strconv.Itoa(id))
	}

	return id, nil
}

func getRequestID(ctx context.Context) string {
	value := ctx.Value(ordermiddleware.RequestIDKey{})
	if value == nil {
		return ""
	}

	if requestID, ok := value.(string); ok {
		return requestID
	}

	return ""
}

func pay(ctx context.Context, orderID int, amount float64) (bool, error) {
	type WithdrawData struct {
		OrderID int     `json:"orderId,omitempty"`
		Amount  float64 `json:"amount,omitempty"`
	}

	data := WithdrawData{
		OrderID: orderID,
		Amount:  amount,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	// todo: Hardcoded
	url := "http://billing.default.svc.cluster.local:8002/account/1/withdraw"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != 200 {
		return false, nil
	}

	return true, nil
}

func sendNotification(event storageorder.OrderEvent, logger Logger, broker MessageBroker) {
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

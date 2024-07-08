package billing

import (
	"context"
	"errors"

	internalbilling "github.com/zaytcevcom/msa/internal/server/billing"
	"github.com/zaytcevcom/msa/internal/storage/billing"
)

type Billing struct {
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
	GetByID(ctx context.Context, id int) *storagebilling.EntityAccount
	Create(ctx context.Context, userID int) (int, error)
	GetBalance(ctx context.Context, id int) (float64, error)
	Deposit(ctx context.Context, accountID int, amount float64) (int, error)
	Withdraw(ctx context.Context, accountID int, orderID int, amount float64) (int, error)
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
}

func New(logger Logger, storage Storage) *Billing {
	return &Billing{
		logger:  logger,
		storage: storage,
	}
}

func (b *Billing) Health(_ context.Context) interface{} {
	return struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}
}

func (b *Billing) CreateAccount(ctx context.Context, userID int) (int, error) {
	id, err := b.storage.Create(ctx, userID)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (b *Billing) GetAccount(ctx context.Context, userID int) (*internalbilling.Account, error) {
	account := b.storage.GetByID(ctx, userID)
	if account == nil {
		return nil, errors.New("account not found")
	}

	balance, err := b.storage.GetBalance(ctx, account.ID)
	if err != nil {
		b.logger.Error(err.Error())
		return nil, errors.New("account balance not found")
	}

	return &internalbilling.Account{
		ID:      account.ID,
		Balance: balance,
	}, nil
}

func (b *Billing) Deposit(ctx context.Context, accountID int, amount float64) (int, error) {
	id, err := b.storage.Deposit(ctx, accountID, amount)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (b *Billing) Withdraw(ctx context.Context, accountID int, orderID int, amount float64) (int, error) {
	balance, err := b.storage.GetBalance(ctx, accountID)
	if err != nil {
		b.logger.Error(err.Error())
		return 0, errors.New("account balance not found")
	}

	if balance-amount < 0 {
		return 0, errors.New("not enough money")
	}

	id, err := b.storage.Withdraw(ctx, accountID, orderID, amount)
	if err != nil {
		return 0, err
	}

	return id, nil
}

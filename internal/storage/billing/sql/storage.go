package sqlstoragebilling

import (
	"context"

	_ "github.com/jackc/pgx/v4/stdlib" //nolint:justifying
	"github.com/jmoiron/sqlx"
	storagebilling "github.com/zaytcevcom/msa/internal/storage/billing"
)

type Storage struct {
	dsn string
	db  *sqlx.DB
}

func New(dsn string) *Storage {
	return &Storage{
		dsn: dsn,
	}
}

func (s *Storage) Connect(ctx context.Context) (err error) {
	s.db, err = sqlx.Open("pgx", s.dsn)
	if err != nil {
		return err
	}

	return s.db.PingContext(ctx)
}

func (s *Storage) Close(_ context.Context) error {
	return s.db.Close()
}

func (s *Storage) GetByID(ctx context.Context, id int) *storagebilling.EntityAccount {
	var accounts []*storagebilling.EntityAccount
	query := "SELECT id, user_id FROM accounts WHERE id = $1"

	err := s.db.SelectContext(ctx, &accounts, query, id)
	if err != nil {
		return nil
	}

	if len(accounts) != 1 {
		return nil
	}

	return accounts[0]
}

func (s *Storage) Create(ctx context.Context, userID int) (int, error) {
	query := `
		INSERT INTO accounts (
			user_id
		) VALUES (
		  $1
	  	)
		RETURNING id
	`
	_, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return 0, err
	}

	var accountID int
	err = s.db.GetContext(ctx, &accountID, "SELECT lastval()")

	return accountID, err
}

func (s *Storage) GetBalance(_ context.Context, id int) (float64, error) {
	query := "SELECT SUM(amount) FROM payments WHERE account_id = $1"

	var balance float64
	err := s.db.QueryRow(query, id).Scan(&balance)
	if err != nil {
		return 0, err
	}

	return balance, nil
}

func (s *Storage) Deposit(ctx context.Context, accountID int, amount float64) (int, error) {
	query := `
		INSERT INTO payments (
			account_id, amount
		) VALUES (
		  $1, $2
	  	)
		RETURNING id
	`
	_, err := s.db.ExecContext(ctx, query, accountID, amount)
	if err != nil {
		return 0, err
	}

	var paymentID int
	err = s.db.GetContext(ctx, &paymentID, "SELECT lastval()")

	return paymentID, err
}

func (s *Storage) Withdraw(ctx context.Context, accountID int, orderID int, amount float64) (int, error) {
	query := `
		INSERT INTO payments (
			account_id, order_id, amount
		) VALUES (
		  $1, $2, $3
	  	)
		RETURNING id
	`
	_, err := s.db.ExecContext(ctx, query, accountID, orderID, -1*amount)
	if err != nil {
		return 0, err
	}

	var paymentID int
	err = s.db.GetContext(ctx, &paymentID, "SELECT lastval()")

	return paymentID, err
}

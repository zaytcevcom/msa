package sqlstorageorder

import (
	"context"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib" //nolint:justifying
	"github.com/jmoiron/sqlx"
	"github.com/zaytcevcom/msa/internal/order"
	storageorder "github.com/zaytcevcom/msa/internal/storage/order"
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

func (s *Storage) GetByID(ctx context.Context, id int) (user *storageorder.EntityOrder) {
	var orders []*storageorder.EntityOrder
	query := "SELECT id, user_id, product_id, sum, status, time FROM orders WHERE id = $1"

	err := s.db.SelectContext(ctx, &orders, query, id)
	if err != nil {
		return nil
	}

	if len(orders) != 1 {
		return nil
	}

	return orders[0]
}

func (s *Storage) Create(
	ctx context.Context,
	userID int,
	productID int,
	sum float64,
	status order.Status,
	time time.Time,
) (int, error) {
	query := `
		INSERT INTO orders (
			user_id,
			product_id,
			sum,
			status,
			time
		) VALUES (
		  $1, $2, $3, $4, $5 
	  	)
		RETURNING id
	`
	_, err := s.db.ExecContext(ctx, query, userID, productID, sum, status, time.Unix())
	if err != nil {
		return 0, err
	}

	var orderID int
	err = s.db.GetContext(ctx, &orderID, "SELECT lastval()")

	return orderID, err
}

func (s *Storage) ChangeStatus(ctx context.Context, orderID int, status order.Status) error {
	query := `
		UPDATE
    		orders
		SET
		    status = $2
		WHERE
		    id = $1
	`
	_, err := s.db.ExecContext(ctx, query, orderID, status)
	return err
}

package sqlstoragestock

import (
	"context"

	_ "github.com/jackc/pgx/v4/stdlib" //nolint:justifying
	"github.com/jmoiron/sqlx"
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

func (s *Storage) GetCountAvailable(_ context.Context, productID int) (int, error) {
	query := `
		SELECT
		    COALESCE((SELECT SUM(p.count) FROM products p WHERE p.id = $1), 0)
		    -
			COALESCE((SELECT SUM(pr.count) FROM product_reserve pr WHERE pr.product_id = $1), 0)
	`

	var count int
	err := s.db.QueryRow(query, productID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Storage) Reserve(ctx context.Context, orderID int, productID int, count int) (int, error) {
	query := `
		INSERT INTO product_reserve (
			order_id,
			product_id,
			count
		) VALUES (
		  $1, $2, $3
	  	)
		RETURNING id
	`
	_, err := s.db.ExecContext(ctx, query, orderID, productID, count)
	if err != nil {
		return 0, err
	}

	var reserveID int
	err = s.db.GetContext(ctx, &reserveID, "SELECT lastval()")

	return reserveID, err
}

func (s *Storage) CancelReserve(ctx context.Context, orderID int) error {
	query := `DELETE FROM product_reserve WHERE order_id = $1`
	_, err := s.db.ExecContext(ctx, query, orderID)
	return err
}

func (s *Storage) RemoveReserve(ctx context.Context, reserveID int) error {
	query := `DELETE FROM product_reserve WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, reserveID)
	return err
}

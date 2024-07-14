package sqlstoragedelivery

import (
	"context"
	"time"

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

func (s *Storage) IsTimeAvailable(_ context.Context, time time.Time) (bool, error) {
	query := `SELECT COUNT(*) FROM employee_reserve WHERE time = $1`

	var count int
	err := s.db.QueryRow(query, time.Unix()).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (s *Storage) Reserve(ctx context.Context, orderID int, time time.Time) (int, error) {
	employeeID := 1

	query := `
		INSERT INTO employee_reserve (
			order_id,
			employee_id,
			time
		) VALUES (
		  $1, $2, $3
	  	)
		RETURNING id
	`
	_, err := s.db.ExecContext(ctx, query, orderID, employeeID, time.Unix())
	if err != nil {
		return 0, err
	}

	var reserveID int
	err = s.db.GetContext(ctx, &reserveID, "SELECT lastval()")

	return reserveID, err
}

func (s *Storage) RemoveReserve(ctx context.Context, reserveID int) error {
	query := `DELETE FROM employee_reserve WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, reserveID)
	return err
}

package sqlstoragenotification

import (
	"context"

	_ "github.com/jackc/pgx/v4/stdlib" //nolint:justifying
	"github.com/jmoiron/sqlx"
	storagenotification "github.com/zaytcevcom/msa/internal/storage/notification"
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

func (s *Storage) GetByUserID(
	ctx context.Context,
	userID int,
) (notifications []storagenotification.EntityNotification) {
	query := `
		SELECT
    		id, email, text
		FROM
			notifications
		WHERE
		    user_id = $1
	`
	err := s.db.SelectContext(ctx, &notifications, query, userID)
	if err != nil {
		return nil
	}
	return notifications
}

func (s *Storage) Create(ctx context.Context, userID int, email string, text string) error {
	query := `
		INSERT INTO notifications (
			user_id, email, text
		) VALUES (
		  $1, $2, $3
	  	)
		RETURNING id
	`
	_, err := s.db.ExecContext(ctx, query, userID, email, text)
	if err != nil {
		return err
	}

	return nil
}

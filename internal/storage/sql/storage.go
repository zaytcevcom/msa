package sqlstorage

import (
	"context"

	_ "github.com/jackc/pgx/v4/stdlib" //nolint:justifying
	"github.com/jmoiron/sqlx"
	"github.com/zaytcevcom/msa/internal/storage"
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

func (s *Storage) GetByID(ctx context.Context, id int) (user *storage.User) {
	var users []*storage.User
	query := "SELECT id, username, first_name, last_name, email, phone FROM users WHERE id = $1"

	err := s.db.SelectContext(ctx, &users, query, id)
	if err != nil {
		return nil
	}

	if len(users) != 1 {
		return nil
	}

	return users[0]
}

func (s *Storage) GetByUsername(ctx context.Context, username string) (user *storage.PasswordDTO) {
	var users []*storage.PasswordDTO
	query := "SELECT id, username, password_hash FROM users WHERE username = $1 LIMIT 1"

	err := s.db.SelectContext(ctx, &users, query, username)
	if err != nil {
		return nil
	}

	if len(users) != 1 {
		return nil
	}

	return users[0]
}

func (s *Storage) Create(ctx context.Context, user storage.UserCreateDTO) (id int, err error) {
	query := `
		INSERT INTO users (
			username, password_hash, first_name, last_name, email, phone
		) VALUES (
		  :username, :password_hash, :first_name, :last_name, :email, :phone
	  	)
		RETURNING id
	`
	_, err = s.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return 0, err
	}

	err = s.db.GetContext(ctx, &id, "SELECT lastval()")

	return id, err
}

func (s *Storage) Update(ctx context.Context, id int, user storage.User) (err error) {
	query := `
		UPDATE
    		users
		SET
		    username = :username,
		    first_name = :first_name,
		    last_name = :last_name,
		    email = :email,
		    phone = :phone
		WHERE
		    id = :id
	`
	user.ID = id
	_, err = s.db.NamedExecContext(ctx, query, user)
	return err
}

func (s *Storage) Delete(ctx context.Context, id int) (err error) {
	query := `DELETE FROM users WHERE id = $1`
	_, err = s.db.ExecContext(ctx, query, id)
	return err
}

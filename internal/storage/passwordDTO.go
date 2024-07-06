package storage

type PasswordDTO struct {
	ID           int
	Username     string
	PasswordHash string `db:"password_hash"`
}

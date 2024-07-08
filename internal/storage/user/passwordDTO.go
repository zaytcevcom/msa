package storageuser

type PasswordDTO struct {
	ID           int
	Username     string
	PasswordHash string `db:"password_hash"`
}

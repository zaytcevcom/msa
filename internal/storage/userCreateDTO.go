package storage

type UserCreateDTO struct {
	Username     string
	PasswordHash string `db:"password_hash"`
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	Email        string
	Phone        string
}

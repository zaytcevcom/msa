package storage

type User struct {
	ID        int
	Username  string
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string
	Phone     string
}

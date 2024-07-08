package storagebilling

type EntityAccount struct {
	ID     int
	UserID int `db:"user_id"`
}

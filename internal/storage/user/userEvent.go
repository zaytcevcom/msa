package storageuser

type UserEvent struct {
	Type   string
	UserID int `db:"user_id"`
}

var UserCreated = "user_created"

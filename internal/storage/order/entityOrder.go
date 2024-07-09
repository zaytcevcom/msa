package storageorder

type EntityOrder struct {
	ID        int
	UserID    int     `db:"user_id"`
	ProductID int     `db:"product_id"`
	Sum       float64 `db:"sum"`
	Status    int     `db:"status"`
	Time      int     `db:"time"`
}

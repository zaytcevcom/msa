package storagestock

type StockReservedEvent struct {
	OrderID int `json:"orderId,omitempty"`
}

type StockNotReservedEvent struct {
	OrderID int `json:"orderId,omitempty"`
}

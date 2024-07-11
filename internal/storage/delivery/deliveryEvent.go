package storagedelivery

type DeliveryReservedEvent struct {
	OrderID int `json:"orderId,omitempty"`
}

type DeliveryNotReservedEvent struct {
	OrderID int `json:"orderId,omitempty"`
}

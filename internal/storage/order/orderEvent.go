package storageorder

type OrderEvent struct {
	Type    string
	OrderID int    `json:"orderId,omitempty"`
	UserID  int    `json:"userId,omitempty"`
	Email   string `json:"email,omitempty"`
	Text    string `json:"text,omitempty"`
}

var OrderCreated = "order_created"

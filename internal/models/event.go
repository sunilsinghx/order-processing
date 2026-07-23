package models

// OrderCreatedEvent represents the payload published when an order is submitted
type OrderCreatedEvent struct {
	EventID  string  `json:"event_id"`
	OrderID  string  `json:"order_id"`
	Customer string  `json:"customer"`
	Amount   float64 `json:"amount"`
}

package models

import "time"

type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusCompleted OrderStatus = "COMPLETED"
	StatusFailed    OrderStatus = "FAILED"
)

type Order struct {
	ID           string      `json:"id"`
	CustomerName string      `json:"customer_name"`
	Amount       float64     `json:"amount"`
	Status       OrderStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
}

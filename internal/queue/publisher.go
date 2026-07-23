package queue

import (
	"context"

	"github.com/sunilsinghx/order-processing/internal/models"
)

// EventPublisher defines the behavior required to emit domain events
type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, event *models.OrderCreatedEvent) error
}

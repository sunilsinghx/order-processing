package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sunilsinghx/order-processing/internal/models"
	"github.com/sunilsinghx/order-processing/internal/queue"
	"github.com/sunilsinghx/order-processing/internal/repository"
)

type OrderService interface {
	CreateOrder(ctx context.Context, name string, amount float64) (*models.Order, error)
	GetOrder(ctx context.Context, id string) (*models.Order, error)
}

type orderService struct {
	repo      repository.OrderRepository
	publisher queue.EventPublisher
}

func NewOrderService(repo repository.OrderRepository, pub queue.EventPublisher) OrderService {
	return &orderService{repo: repo, publisher: pub}
}

func (s *orderService) CreateOrder(ctx context.Context, name string, amount float64) (*models.Order, error) {
	order := &models.Order{
		ID:           string(uuid.New().String())[:6],
		CustomerName: name,
		Amount:       amount,
		Status:       models.StatusPending,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	// create OrderCreatedEvent
	event := &models.OrderCreatedEvent{
		EventID:  string(uuid.New().String())[:6],
		OrderID:  order.ID,
		Customer: order.CustomerName,
		Amount:   order.Amount,
	}

	// Emit asynchronously out to the message broker
	if err := s.publisher.PublishOrderCreated(ctx, event); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	return s.repo.GetByID(ctx, id)
}

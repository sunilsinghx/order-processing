package queue

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sunilsinghx/order-processing/internal/models"
)

func ConnectRabbitMQ(url string) (*amqp091.Connection, error) {
	var conn *amqp091.Connection
	var err error

	for i := 1; i <= 10; i++ {
		conn, err = amqp091.Dial(url)
		if err == nil {
			return conn, nil
		}

		log.Printf("RabbitMQ not ready (%d/10): %v", i, err)
		time.Sleep(2 * time.Second)
	}

	return nil, err
}

type RabbitMQPublisher struct {
	channel  *amqp.Channel
	exchange string
}

// NewRabbitMQPublisher constructs a publisher matching the EventPublisher interface
func NewRabbitMQPublisher(ch *amqp.Channel, exchange string) EventPublisher {
	return &RabbitMQPublisher{
		channel:  ch,
		exchange: exchange,
	}
}

func (p *RabbitMQPublisher) PublishOrderCreated(ctx context.Context, event *models.OrderCreatedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Publish the message using standard AMQP routing properties
	return p.channel.PublishWithContext(ctx,
		p.exchange,
		"order.created",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

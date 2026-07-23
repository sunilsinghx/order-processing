package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sunilsinghx/order-processing/internal/metrics"
	"github.com/sunilsinghx/order-processing/internal/models"
	"github.com/sunilsinghx/order-processing/internal/repository"
)

type OrderWorker struct {
	ch          *amqp.Channel
	orderRepo   repository.OrderRepository
	idempotRepo repository.IdempotencyRepository
}

func NewOrderWorker(
	ch *amqp.Channel,
	orderRepo repository.OrderRepository,
	idempotRepo repository.IdempotencyRepository,
) *OrderWorker {
	return &OrderWorker{
		ch:          ch,
		orderRepo:   orderRepo,
		idempotRepo: idempotRepo,
	}
}

func (w *OrderWorker) Start(ctx context.Context, queueName string, concurrency int) error {
	err := w.ch.Qos(concurrency, 0, false)
	if err != nil {
		return err
	}

	msgs, err := w.ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i := 1; i <= concurrency; i++ {
		wg.Add(1)
		go w.workerConsumer(ctx, &wg, i, msgs)
	}

	<-ctx.Done()
	wg.Wait()
	return nil
}

func (w *OrderWorker) workerConsumer(ctx context.Context, wg *sync.WaitGroup, workerID int, msgs <-chan amqp.Delivery) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			w.processMessage(ctx, workerID, d)
		}
	}
}

func (w *OrderWorker) processMessage(ctx context.Context, workerID int, d amqp.Delivery) {

	metrics.WorkerBusy.Inc()
	defer metrics.WorkerBusy.Dec()

	status := "success"
	start := time.Now()
	defer func() {
		metrics.ProcessingDuration.
			WithLabelValues("orders.processing_worker_queue", status).
			Observe(time.Since(start).Seconds())
	}()

	var event models.OrderCreatedEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Printf("[Worker #%d ERROR] Malformed event payload: %v", workerID, err)
		// No requeue, no retry: instantly send malformed payloads to DLQ

		d.Nack(false, false)
		w.orderRepo.UpdateStatus(ctx, event.OrderID, models.StatusFailed)
		return
	}

	// 1. Check and claim Idempotency
	err := w.idempotRepo.ClaimEvent(ctx, event.EventID)
	if err != nil {
		if errors.Is(err, repository.ErrEventAlreadyProcessed) {
			log.Printf("[Worker #%d IDEMPOTENCY] Event %s already executed. Acknowledging.", workerID, event.EventID)
			d.Ack(false)
			return
		}
		log.Printf("[Worker #%d ERROR] Idempotency DB check failed: %v", workerID, err)
		d.Nack(false, true)
		status = "failed"
		return
	}

	// 2. Processing with Exponential Backoff Retry
	const maxRetries = 3
	const initialBackoff = 2 * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {

		log.Printf("[Worker #%d] Processing Event %s (Attempt %d/%d)...", workerID, event.EventID, attempt, maxRetries)

		lastErr = w.executeBusinessWorkload(ctx, &event)
		if lastErr == nil {
			break
		}

		log.Printf("[Worker #%d] Attempt %d failed: %v", workerID, attempt, lastErr)

		if attempt < maxRetries {
			// Exponential Backoff: initial * (2^(attempt-1)) -> 2s, 4s, 8s...
			backoffDuration := initialBackoff * time.Duration(math.Pow(2, float64(attempt-1)))
			log.Printf("[Worker #%d] Backing off. Retrying in %v...", workerID, backoffDuration)

			select {
			case <-ctx.Done():
				// If the application is shutting down, stop waiting and requeue
				d.Nack(false, true)
				metrics.OrdersFailed.Inc()
				metrics.ProcessedMessages.WithLabelValues("failed").Inc()

				return
			case <-time.After(backoffDuration):
			}
		}
	}

	// 3. Evaluate Final Processing Results
	if lastErr != nil {
		log.Printf("[Worker #%d CRITICAL] Max retries exhausted for Event %s. Routing to DLQ.", workerID, event.EventID)

		// Crucial Step: Nack with requeue=false.
		// Because the queue is declared with "x-dead-letter-exchange",
		// RabbitMQ immediately routes this message to "orders.dlq"
		d.Nack(false, false)
		metrics.OrdersFailed.Inc()
		metrics.ProcessedMessages.WithLabelValues("failed").Inc()

		w.orderRepo.UpdateStatus(ctx, event.OrderID, models.StatusFailed)
		return
	}

	metrics.OrdersCompleted.Inc()
	metrics.ProcessedMessages.WithLabelValues("success").Inc()
	log.Printf("[Worker #%d SUCCESS] Finalized execution for Order ID: %s", workerID, event.OrderID)
	d.Ack(false)
}

func (w *OrderWorker) executeBusinessWorkload(ctx context.Context, event *models.OrderCreatedEvent) error {
	// Any Order with an amount divisible by 10 will fail to simulate the retry mechanism
	if int(event.Amount)%10 == 0 {
		return fmt.Errorf("inventory validation failed: service is currently unavailable")
	}

	time.Sleep(3 * time.Second)
	// If successful, update the database order status
	return w.orderRepo.UpdateStatus(ctx, event.OrderID, models.StatusCompleted)
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sunilsinghx/order-processing/internal/config"
	"github.com/sunilsinghx/order-processing/internal/db"
	"github.com/sunilsinghx/order-processing/internal/metrics"
	"github.com/sunilsinghx/order-processing/internal/queue"
	"github.com/sunilsinghx/order-processing/internal/repository"
	"github.com/sunilsinghx/order-processing/internal/worker"
)

func main() {

	// Metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("[INFO] Worker metrics server running on :8081/metrics")

		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatalf("Metrics server failed: %v", err)
		}
	}()

	cfg := config.Load()

	// Database
	db, err := db.New(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close(context.Background())

	orderRepo := repository.NewPostgresRepository(db)
	idempotRepo := repository.NewPostgresIdempotencyRepository(db)

	var (
		conn *amqp091.Connection
	)

	//connect rabbitmq
	conn, err = queue.ConnectRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after retries: %v", err)
	}
	defer conn.Close()

	// Main worker channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to create channel: %v", err)
	}
	defer ch.Close()

	// EXCHANGES

	// Main exchange
	err = ch.ExchangeDeclare(
		"orders.exchange",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to declare orders.exchange: %v", err)
	}

	// Dead letter exchange
	err = ch.ExchangeDeclare(
		"orders.dlx",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to declare orders.dlx: %v", err)
	}

	// DEAD LETTER QUEUE

	dlq, err := ch.QueueDeclare(
		"orders.dlq",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to declare DLQ: %v", err)
	}

	err = ch.QueueBind(
		dlq.Name,
		"order.failed",
		"orders.dlx",
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to bind DLQ: %v", err)
	}

	// MAIN QUEUE

	mainQueueArgs := amqp.Table{
		"x-dead-letter-exchange":    "orders.dlx",
		"x-dead-letter-routing-key": "order.failed",
	}

	queueName := "orders.processing_worker_queue"

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		mainQueueArgs,
	)

	if err != nil {
		log.Fatalf("Failed to declare worker queue: %v", err)
	}

	err = ch.QueueBind(
		q.Name,
		"order.created",
		"orders.exchange",
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to bind worker queue: %v", err)
	}

	// QUEUE METRICS CHANNEL
	metricsCh, err := conn.Channel()

	if err != nil {
		log.Fatalf("Failed to create metrics channel: %v", err)
	}

	defer metricsCh.Close()

	go func() {

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {

			queue, err := metricsCh.QueueInspect(queueName)

			if err != nil {
				log.Printf("Queue inspect failed: %v", err)
				continue
			}

			metrics.QueueDepth.
				WithLabelValues(queueName).
				Set(float64(queue.Messages))
		}

	}()

	// START WORKER
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orderWorker := worker.NewOrderWorker(
		ch,
		orderRepo,
		idempotRepo,
	)

	go func() {

		concurrency, _ := strconv.Atoi(cfg.GoRoutine)
		err := orderWorker.Start(
			ctx,
			q.Name,
			concurrency,
		)

		if err != nil {
			log.Printf("[FATAL] Worker stopped: %v", err)
			cancel()
		}

	}()

	// Shutdown handling
	stopChan := make(chan os.Signal, 1)

	signal.Notify(
		stopChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-stopChan

	log.Println("[SHUTDOWN] Worker stopping...")
}

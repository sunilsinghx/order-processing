package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/sunilsinghx/order-processing/internal/config"
	_ "github.com/sunilsinghx/order-processing/internal/config"
	"github.com/sunilsinghx/order-processing/internal/db"
	_ "github.com/sunilsinghx/order-processing/internal/db"
	"github.com/sunilsinghx/order-processing/internal/handler"

	"github.com/sunilsinghx/order-processing/internal/middleware"
	"github.com/sunilsinghx/order-processing/internal/queue"
	"github.com/sunilsinghx/order-processing/internal/repository"
	"github.com/sunilsinghx/order-processing/internal/service"
)

func main() {

	cfg := config.Load()
	var (
		conn *amqp091.Connection
		err  error
	)

	//connect rabbitmq
	conn, err = queue.ConnectRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after retries: %v", err)
	}
	defer conn.Close()

	db, err := db.New(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close(context.Background())

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare the exchange where events get routed
	err = ch.ExchangeDeclare("orders.exchange", "topic", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()

	// Wire dependencies
	repo := repository.NewPostgresRepository(db)
	pub := queue.NewRabbitMQPublisher(ch, "orders.exchange")
	svc := service.NewOrderService(repo, pub)
	orderHandler := handler.NewOrderHandler(svc)
	// Set up Routing
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	// Attach Global Observability Middleware
	r.Use(middleware.TrackMetrics())

	// Public Scraper Endpoint for Prometheus Integration
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Application Protected Endpoint with Rate Limiting (10 req / minute)
	limiter := middleware.NewRateLimiter(rdb, 10, time.Minute)

	api := r.Group("/api/v1")
	api.Use(limiter.LimitByIP())
	{
		api.POST("/orders", orderHandler.Create)
	}

	r.GET("/orders/:id", orderHandler.GetByID)

	r.Run(":8080")
}

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// API Metrics
var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed.",
		},
		[]string{"path", "method", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latency of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	RateLimitTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limit_total",
			Help: "Total number of requests rejected by rate limiting.",
		},
	)
)

// Worker Metrics
var (
	QueueDepth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_depth",
			Help: "Current number of messages waiting in the queue.",
		},
		[]string{"queue_name"},
	)

	ProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "processing_duration_seconds",
			Help:    "Time taken to process a single queued message.",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"queue_name", "status"},
	)

	WorkerBusy = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "worker_busy_goroutines",
			Help: "Current number of active worker goroutines processing jobs.",
		},
	)

	ProcessedMessages = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "processed_messages_total",
			Help: "Total processed messages.",
		},
		[]string{"status"},
	)
	OrdersCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Orders successfully created.",
		},
	)

	OrdersFailed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_failed_total",
			Help: "Orders failed.",
		},
	)

	OrdersCompleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_completed_total",
			Help: "Orders completed.",
		},
	)
)

# --- Stage 1: Build binary ---
    FROM golang:1.25-alpine AS builder

    WORKDIR /app
    
    # Copy dependency files first
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the rest of the application code
    COPY . .
    
    # Build the Worker binary static and optimized
    RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o worker-service ./cmd/worker
    
    # --- Stage 2: Final lightweight image ---
    FROM alpine:3.21
    
    WORKDIR /app
    
    
    # Copy the binary from the builder stage
    COPY --from=builder /app/worker-service .
    
    # Expose the metrics port (8081) so Prometheus can scrape it
    EXPOSE 8081
    
    # Run the Worker application
    CMD ["./worker-service"]
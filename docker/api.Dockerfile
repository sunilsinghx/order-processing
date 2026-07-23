# --- Stage 1: Build binary ---
    FROM golang:1.25-alpine AS builder

    WORKDIR /app
    
    # Copy dependency files first to utilize Docker layer caching
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the rest of the application code
    COPY . .
    
    # Build the API binary static and optimized
    RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api-service ./cmd/api
    
    # --- Stage 2: Final lightweight image ---
    FROM alpine:3.21
    
    WORKDIR /app
    
    
    # Copy the binary from the builder stage
    COPY --from=builder /app/api-service .
    
    # Expose the API port
    EXPOSE 8080
    
    # Run the API application
    CMD ["./api-service"]
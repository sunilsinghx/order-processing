-- Orders table
CREATE TABLE orders (
    id TEXT PRIMARY KEY,
    customer_name VARCHAR(255) NOT NULL,
    amount NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT orders_status_check
    CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED'))
);


-- Idempotency table
CREATE TABLE processed_events (
    event_id TEXT PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_orders_status
ON orders(status);

CREATE INDEX idx_orders_created_at
ON orders(created_at DESC);
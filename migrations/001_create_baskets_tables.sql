-- migrations/001_create_baskets_tables.sql

-- Enable UUID generation if not already enabled
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Table for Baskets
CREATE TABLE IF NOT EXISTS baskets (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- Add user_id UUID REFERENCES users(id) later if needed
);

-- Table for Items (Stocks) within a Basket
CREATE TABLE IF NOT EXISTS basket_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Or use SERIAL PRIMARY KEY if preferred
    basket_id UUID NOT NULL REFERENCES baskets(id) ON DELETE CASCADE, -- Link to basket
    symbol VARCHAR(50) NOT NULL,
    quantity INT NOT NULL CHECK (quantity > 0), -- Ensure positive quantity
    UNIQUE (basket_id, symbol) -- Prevent adding the same stock twice to one basket
);

-- Optional: Indexes for performance
CREATE INDEX IF NOT EXISTS idx_basket_items_basket_id ON basket_items(basket_id);

-- Optional: Trigger to update baskets.updated_at timestamp automatically
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_baskets_updated_at
BEFORE UPDATE ON baskets
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
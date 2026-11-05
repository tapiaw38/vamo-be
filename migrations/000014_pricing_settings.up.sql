CREATE TABLE pricing_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    base_price NUMERIC(14,2) NOT NULL,
    price_per_km NUMERIC(14,2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'ARS',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by TEXT
);

INSERT INTO pricing_settings (id, base_price, price_per_km, currency)
VALUES (1, 2000.00, 150.00, 'ARS')
ON CONFLICT (id) DO NOTHING;

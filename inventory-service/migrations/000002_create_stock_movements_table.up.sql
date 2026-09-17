CREATE TABLE stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lot_id UUID NOT NULL REFERENCES lots(id),
    type TEXT NOT NULL,
    quantity_kg NUMERIC(12,4) NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    origin TEXT NOT NULL
);

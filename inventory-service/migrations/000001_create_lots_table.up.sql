CREATE TABLE lots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    purchase_item_id UUID NOT NULL,
    safra TEXT NOT NULL,
    quantity_kg NUMERIC(12,4) NOT NULL,
    received_at DATE NOT NULL
);

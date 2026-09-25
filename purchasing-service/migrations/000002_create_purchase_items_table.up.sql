CREATE TABLE purchase_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_id UUID NOT NULL REFERENCES purchases(id),
    product_id UUID NOT NULL,
    unit_id UUID NOT NULL,
    quantity NUMERIC(12,4) NOT NULL
);

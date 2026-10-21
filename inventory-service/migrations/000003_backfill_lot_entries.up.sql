INSERT INTO stock_movements (lot_id, type, quantity_kg, occurred_at, origin)
SELECT l.id, 'entrada', l.quantity_kg, l.received_at, 'compra (migracao 000003)'
FROM lots l
WHERE NOT EXISTS (
    SELECT 1 FROM stock_movements m WHERE m.lot_id = l.id
);

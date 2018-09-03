ALTER TABLE white_water_rapid ADD COLUMN order_index INTEGER NOT NULL DEFAULT 0;
ALTER TABLE white_water_rapid ADD COLUMN auto_ordering BOOLEAN NOT NULL DEFAULT TRUE;
CREATE INDEX white_water_rapid_order_index ON white_water_rapid(order_index);

ALTER TABLE white_water_rapid ADD COLUMN order_index INTEGER NOT NULL DEFAULT 0;
ALTER TABLE white_water_rapid ADD COLUMN manual_change BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX white_water_rapid_order_index ON white_water_rapid(order_index);

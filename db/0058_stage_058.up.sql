ALTER TABLE transfer_river
    DROP CONSTRAINT transfer_river_transfer_id_fkey;

ALTER TABLE transfer_river
    ADD CONSTRAINT transfer_river_transfer_id_fkey
        FOREIGN KEY (transfer_id) REFERENCES transfer
            ON DELETE CASCADE;

ALTER TABLE transfer_river
    DROP CONSTRAINT transfer_river_river_id_fkey;

ALTER TABLE transfer_river
    ADD CONSTRAINT transfer_river_river_id_fkey
        FOREIGN KEY (river_id) REFERENCES river
            ON DELETE CASCADE;
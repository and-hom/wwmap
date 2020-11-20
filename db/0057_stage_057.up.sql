UPDATE voyage_report
SET date_modified=NOW()
WHERE date_modified IS NULL;

ALTER TABLE voyage_report
    ALTER COLUMN date_modified
        SET NOT NULL;


ALTER TABLE voyage_report_river
    DROP CONSTRAINT voyage_report_river_voyage_report_id_fkey;

ALTER TABLE voyage_report_river
    ADD CONSTRAINT voyage_report_river_voyage_report_id_fkey
        FOREIGN KEY (voyage_report_id) REFERENCES voyage_report
            ON DELETE CASCADE;

ALTER TABLE voyage_report_river
    DROP CONSTRAINT voyage_report_river_river_id_fkey;

ALTER TABLE voyage_report_river
    ADD CONSTRAINT voyage_report_river_river_id_fkey
        FOREIGN KEY (river_id) REFERENCES river
            ON DELETE CASCADE;
 
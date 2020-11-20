ALTER TABLE voyage_report
    ALTER COLUMN date_modified
        DROP NOT NULL;

ALTER TABLE voyage_report_river
    DROP CONSTRAINT voyage_report_river_voyage_report_id_fkey;

ALTER TABLE voyage_report_river
    ADD CONSTRAINT voyage_report_river_voyage_report_id_fkey
        FOREIGN KEY (voyage_report_id) REFERENCES voyage_report;

ALTER TABLE voyage_report_river
    DROP CONSTRAINT voyage_report_river_river_id_fkey;

ALTER TABLE voyage_report_river
    ADD CONSTRAINT voyage_report_river_river_id_fkey
        FOREIGN KEY (river_id) REFERENCES river;

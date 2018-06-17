DELETE FROM image;
ALTER TABLE image ADD COLUMN report_id BIGINT NOT NULL REFERENCES voyage_report(id);

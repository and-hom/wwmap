ALTER TABLE voyage_report ADD COLUMN author CHARACTER VARYING(512);

-- reimport reports
UPDATE voyage_report SET date_modified=(SELECT min(date_of_trip) FROM voyage_report);
UPDATE voyage_report SET title=SUBSTRING(title, 1, 1024);
ALTER TABLE voyage_report ALTER COLUMN title TYPE CHARACTER VARYING(1024);
ALTER TABLE voyage_report DROP COLUMN date_of_trip;
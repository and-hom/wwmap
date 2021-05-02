DELETE FROM voyage_report WHERE removed=TRUE and source='MANUAL';
ALTER TABLE voyage_report DROP COLUMN removed;

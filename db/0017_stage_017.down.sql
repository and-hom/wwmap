DROP TRIGGER IF EXISTS river_last_modified_trigger ON river;
DROP FUNCTION IF EXISTS set_last_modified_trigger_function();
ALTER TABLE river DROP COLUMN last_modified;
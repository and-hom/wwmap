ALTER TABLE river ADD COLUMN last_modified TIMESTAMP NOT NULL DEFAULT now();

CREATE OR REPLACE FUNCTION set_last_modified_trigger_function()
RETURNS trigger AS '
BEGIN
  NEW.last_modified = now();
  RETURN NEW;
END' LANGUAGE 'plpgsql';

CREATE TRIGGER river_last_modified_trigger
BEFORE UPDATE ON river
FOR EACH ROW
EXECUTE PROCEDURE set_last_modified_trigger_function();

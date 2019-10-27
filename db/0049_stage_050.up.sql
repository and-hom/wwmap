UPDATE white_water_rapid SET aliases='[]' WHERE aliases='null';
ALTER TABLE white_water_rapid ADD CONSTRAINT aliases_is_not_null CHECK ( aliases != 'null' );

UPDATE river SET aliases='[]' WHERE aliases='null';
ALTER TABLE river ADD CONSTRAINT aliases_is_not_null CHECK ( aliases != 'null' );

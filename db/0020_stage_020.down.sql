ALTER TABLE river ALTER COLUMN region_id DROP NOT NULL;
UPDATE river SET region_id = NULL WHERE region_id = (SELECT id FROM region WHERE title='-');
DELETE FROM region WHERE title='-';
DELETE FROM country WHERE title='-';
INSERT INTO country(title) VALUES ('-');
INSERT INTO region(title, country_id, fake) VALUES ('-', (SELECT id FROM country WHERE title='-'), TRUE);
UPDATE river SET region_id = (SELECT id FROM region WHERE title='-') WHERE region_id IS NULL;

ALTER TABLE river ALTER COLUMN region_id SET NOT NULL;
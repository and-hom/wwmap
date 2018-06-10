CREATE TABLE country(
  id                BIGINT PRIMARY KEY                DEFAULT nextval('id_gen'),
  title             CHARACTER VARYING(128)            UNIQUE NOT NULL
);

CREATE TABLE region(
  id                BIGINT PRIMARY KEY	              DEFAULT nextval('id_gen'),
  country_id        BIGINT                            NOT NULL REFERENCES country(id),
  title             CHARACTER VARYING(128)            UNIQUE NOT NULL,
  fake              BOOL
);
CREATE INDEX region_country_id ON region(country_id);

ALTER TABLE river ADD COLUMN region_id BIGINT REFERENCES region(id);
CREATE INDEX river_region_id ON river(region_id);


INSERT INTO country(title) VALUES ('Россия'),('Абхазия');

WITH
    russia AS (SELECT id FROM country WHERE title='Россия'),
    abkhasia AS (SELECT id FROM country WHERE title='Абхазия')
INSERT INTO region(title, country_id) VALUES
    ('Кольский', (select id from russia)),
    ('Карелия', (select id from russia)),
    ('Южный Урал', (select id from russia)),
    ('Полярный Урал', (select id from russia)),
    ('Северный Кавказ', (select id from russia)),
    ('Прибайкалье', (select id from russia)),
    ('Алтай', (select id from russia)),
    ('Абхазия', (select id from abkhasia));
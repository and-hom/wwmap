ALTER TABLE country ADD COLUMN code CHARACTER VARYING(3) NOT NULL DEFAULT '-';
UPDATE country SET code='RU' WHERE title='Россия';
UPDATE country SET code='AB' WHERE title='Абхазия';

ALTER TABLE river ADD COLUMN description TEXT NOT NULL DEFAULT '';
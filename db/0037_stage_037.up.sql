ALTER TABLE report RENAME TO "notification";
ALTER TABLE "notification" ADD COLUMN provider CHARACTER VARYING(16) NOT NULL DEFAULT 'email';
ALTER TABLE "notification" ADD COLUMN recipient CHARACTER VARYING(64) NOT NULL DEFAULT 'info@wwmap.ru';

ALTER TABLE "notification" ADD COLUMN title CHARACTER VARYING(512) NOT NULL DEFAULT '';
ALTER TABLE "notification" ADD COLUMN object_title CHARACTER VARYING(512) NOT NULL DEFAULT '';

UPDATE "notification" SET
    title=COALESCE((SELECT title FROM river WHERE river.id=(SELECT id FROM white_water_rapid WHERE white_water_rapid.id=object_id)),'-'),
    object_title=COALESCE((SELECT title FROM white_water_rapid WHERE white_water_rapid.id=object_id),'-');


ALTER TABLE "notification" ADD COLUMN classifier CHARACTER VARYING(8) NOT NULL DEFAULT 'wwmap';
ALTER TABLE "notification" ADD COLUMN send_before timestamp NOT NULL DEFAULT now();

CREATE INDEX notification_classifier_provider_recipient
  ON "notification"(provider,recipient,classifier);

DROP INDEX notification_classifier_provider_recipient;
ALTER TABLE "notification" DROP COLUMN provider;
ALTER TABLE "notification" DROP COLUMN recipient;
ALTER TABLE "notification" DROP COLUMN title;
ALTER TABLE "notification" DROP COLUMN object_title;
ALTER TABLE "notification" DROP COLUMN classifier;
ALTER TABLE "notification" DROP COLUMN send_before;
ALTER TABLE "notification" RENAME TO "report";
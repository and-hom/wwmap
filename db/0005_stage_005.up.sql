ALTER TABLE track
  ADD COLUMN "start_time" TIMESTAMP NOT NULL DEFAULT now();
ALTER TABLE track
  ADD COLUMN "end_time" TIMESTAMP NOT NULL DEFAULT now();

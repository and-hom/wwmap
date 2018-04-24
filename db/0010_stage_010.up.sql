ALTER TABLE report ADD COLUMN created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now();
ALTER TABLE report ADD COLUMN "read" BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX report_created_at ON report USING btree (created_at);
CREATE INDEX report_read ON report USING btree ("read");

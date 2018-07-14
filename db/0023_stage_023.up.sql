INSERT INTO voyage_report(id,title,source,remote_id,url) VALUES(0,'-','wwmap','0','')
ON CONFLICT DO NOTHING;

ALTER TABLE image ADD COLUMN enabled BOOL NOT NULL DEFAULT TRUE;
ALTER TABLE image ADD COLUMN "type" CHARACTER VARYING(8) NOT NULL DEFAULT 'image';
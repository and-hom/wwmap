-- preserve old val
ALTER TABLE "user" ADD COLUMN session_id CHARACTER VARYING(128);
CREATE UNIQUE INDEX user_session_id ON "user"(session_id);
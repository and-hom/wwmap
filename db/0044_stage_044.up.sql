CREATE TABLE changes_log
(
  id            BIGINT PRIMARY KEY DEFAULT nextval('id_gen'),
  object_type   CHARACTER VARYING(16) NOT NULL,
  object_id     BIGINT NOT NULL,
  auth_provider CHARACTER VARYING(16) NOT NULL,
  ext_id        CHARACTER VARYING(128) NOT NULL,
  login        CHARACTER VARYING(128) NOT NULL,
  "type"        CHARACTER VARYING(16) NOT NULL,
  description   TEXT NOT NULL,
  "time"        TIMESTAMP  NOT NULL DEFAULT now()
);

CREATE INDEX changes_log_object_type_object_id_idx ON changes_log(object_type, object_id);
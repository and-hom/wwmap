CREATE TRIGGER waterway_path_simplified_trigger
    BEFORE UPDATE OF "path" OR INSERT
    ON waterway
    FOR EACH ROW
EXECUTE PROCEDURE set_waterway_path_simplified();

CREATE TRIGGER waterway_path_simplified_change_trigger
    BEFORE UPDATE OF "path_simplified"
    ON waterway
    FOR EACH ROW
EXECUTE PROCEDURE path_simplified_changed();

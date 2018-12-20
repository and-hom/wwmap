CREATE OR REPLACE FUNCTION wwmap_search(txt character varying, queries character varying[]) RETURNS integer AS $$
DECLARE
  counter integer := 0;
  q character varying;
BEGIN
  txt = replace(txt, 'ё', 'е');
  FOREACH q IN ARRAY queries LOOP
    IF txt ilike '%'||q||'%' THEN
      counter = counter + 1;
    END IF;
  END LOOP;
  RETURN counter;
END
$$ LANGUAGE plpgsql;

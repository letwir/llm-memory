-- Application-owned WORM boundary: raw facts and lineage are immutable.
-- Lifecycle transitions and derived compaction fields remain writable by the CLI.
CREATE OR REPLACE FUNCTION llm_memory_guard_immutable_fields()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'memories are append-only; DELETE is forbidden (id=%)', OLD.id;
    END IF;

    IF NEW.id IS DISTINCT FROM OLD.id
       OR NEW.client_id IS DISTINCT FROM OLD.client_id
       OR NEW.category IS DISTINCT FROM OLD.category
       OR NEW.title IS DISTINCT FROM OLD.title
       OR NEW.content_l0 IS DISTINCT FROM OLD.content_l0
       OR NEW.version IS DISTINCT FROM OLD.version
       OR NEW.valid_from IS DISTINCT FROM OLD.valid_from
       OR NEW.tx_created_at IS DISTINCT FROM OLD.tx_created_at
       OR NEW.created_at IS DISTINCT FROM OLD.created_at THEN
        RAISE EXCEPTION 'immutable memory fields cannot be changed (id=%)', OLD.id;
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS memories_immutable_fields_trigger ON memories;
CREATE TRIGGER memories_immutable_fields_trigger
BEFORE UPDATE OR DELETE ON memories
FOR EACH ROW
EXECUTE FUNCTION llm_memory_guard_immutable_fields();

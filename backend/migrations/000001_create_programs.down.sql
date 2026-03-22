DROP TRIGGER IF EXISTS trg_programs_updated_at ON programs;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_programs_name;
DROP INDEX IF EXISTS idx_programs_status;
DROP TABLE IF EXISTS programs;

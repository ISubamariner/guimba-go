-- Standardization migration:
-- 1. Add explicit FK constraint names (fk_<table>_<column> convention)
-- 2. Standardize UUID generation to gen_random_uuid() (PostgreSQL 13+ built-in)
-- 3. Add missing updated_at column to permissions table

-- ============================================================
-- 1. Add explicit FK constraint names
-- ============================================================

-- role_permissions: rename anonymous FKs
ALTER TABLE role_permissions
    DROP CONSTRAINT IF EXISTS role_permissions_role_id_fkey,
    ADD CONSTRAINT fk_role_permissions_role_id
        FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;

ALTER TABLE role_permissions
    DROP CONSTRAINT IF EXISTS role_permissions_permission_id_fkey,
    ADD CONSTRAINT fk_role_permissions_permission_id
        FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE;

-- user_roles: rename anonymous FKs
ALTER TABLE user_roles
    DROP CONSTRAINT IF EXISTS user_roles_user_id_fkey,
    ADD CONSTRAINT fk_user_roles_user_id
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE user_roles
    DROP CONSTRAINT IF EXISTS user_roles_role_id_fkey,
    ADD CONSTRAINT fk_user_roles_role_id
        FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;

-- program_beneficiaries: rename anonymous FKs
ALTER TABLE program_beneficiaries
    DROP CONSTRAINT IF EXISTS program_beneficiaries_program_id_fkey,
    ADD CONSTRAINT fk_program_beneficiaries_program_id
        FOREIGN KEY (program_id) REFERENCES programs(id) ON DELETE CASCADE;

ALTER TABLE program_beneficiaries
    DROP CONSTRAINT IF EXISTS program_beneficiaries_beneficiary_id_fkey,
    ADD CONSTRAINT fk_program_beneficiaries_beneficiary_id
        FOREIGN KEY (beneficiary_id) REFERENCES beneficiaries(id) ON DELETE CASCADE;

-- ============================================================
-- 2. Standardize UUID generation to gen_random_uuid()
-- ============================================================

-- programs: switch from uuid_generate_v4() to gen_random_uuid()
ALTER TABLE programs ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- roles: switch from uuid_generate_v4() to gen_random_uuid()
ALTER TABLE roles ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- permissions: switch from uuid_generate_v4() to gen_random_uuid()
ALTER TABLE permissions ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- users: switch from uuid_generate_v4() to gen_random_uuid()
ALTER TABLE users ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- (beneficiaries and program_beneficiaries already use gen_random_uuid())

-- ============================================================
-- 3. Add missing updated_at column to permissions table
-- ============================================================

ALTER TABLE permissions
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE TRIGGER trg_permissions_updated_at
    BEFORE UPDATE ON permissions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

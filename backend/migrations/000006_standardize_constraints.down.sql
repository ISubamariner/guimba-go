-- Reverse standardization migration

-- ============================================================
-- 3. Remove updated_at from permissions
-- ============================================================

DROP TRIGGER IF EXISTS trg_permissions_updated_at ON permissions;
ALTER TABLE permissions DROP COLUMN IF EXISTS updated_at;

-- ============================================================
-- 2. Revert UUID generation to uuid_generate_v4()
-- ============================================================

ALTER TABLE programs ALTER COLUMN id SET DEFAULT uuid_generate_v4();
ALTER TABLE roles ALTER COLUMN id SET DEFAULT uuid_generate_v4();
ALTER TABLE permissions ALTER COLUMN id SET DEFAULT uuid_generate_v4();
ALTER TABLE users ALTER COLUMN id SET DEFAULT uuid_generate_v4();

-- ============================================================
-- 1. Revert FK constraint names to anonymous defaults
-- ============================================================

-- program_beneficiaries
ALTER TABLE program_beneficiaries
    DROP CONSTRAINT IF EXISTS fk_program_beneficiaries_beneficiary_id,
    ADD FOREIGN KEY (beneficiary_id) REFERENCES beneficiaries(id) ON DELETE CASCADE;

ALTER TABLE program_beneficiaries
    DROP CONSTRAINT IF EXISTS fk_program_beneficiaries_program_id,
    ADD FOREIGN KEY (program_id) REFERENCES programs(id) ON DELETE CASCADE;

-- user_roles
ALTER TABLE user_roles
    DROP CONSTRAINT IF EXISTS fk_user_roles_role_id,
    ADD FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;

ALTER TABLE user_roles
    DROP CONSTRAINT IF EXISTS fk_user_roles_user_id,
    ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- role_permissions
ALTER TABLE role_permissions
    DROP CONSTRAINT IF EXISTS fk_role_permissions_permission_id,
    ADD FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE;

ALTER TABLE role_permissions
    DROP CONSTRAINT IF EXISTS fk_role_permissions_role_id,
    ADD FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;

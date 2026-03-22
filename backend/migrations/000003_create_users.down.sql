DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP INDEX IF EXISTS idx_user_roles_role;
DROP INDEX IF EXISTS idx_user_roles_user;
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;

-- Remove seeded data (in reverse order of dependencies)
DELETE FROM role_permissions WHERE role_id IN (SELECT id FROM roles WHERE is_system_role = TRUE);
DELETE FROM permissions WHERE is_system_permission = TRUE;
DELETE FROM roles WHERE is_system_role = TRUE;

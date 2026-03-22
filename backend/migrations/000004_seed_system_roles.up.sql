-- Seed system roles and permissions

-- System roles
INSERT INTO roles (name, display_name, description, is_system_role) VALUES
    ('admin', 'Administrator', 'Full system access', TRUE),
    ('staff', 'Staff', 'Can manage programs and beneficiaries', TRUE),
    ('viewer', 'Viewer', 'Read-only access', TRUE);

-- Permissions by category
INSERT INTO permissions (name, display_name, category, is_system_permission) VALUES
    -- Programs
    ('programs.create', 'Create Programs', 'programs', TRUE),
    ('programs.read', 'View Programs', 'programs', TRUE),
    ('programs.update', 'Update Programs', 'programs', TRUE),
    ('programs.delete', 'Delete Programs', 'programs', TRUE),
    -- Users
    ('users.create', 'Create Users', 'users', TRUE),
    ('users.read', 'View Users', 'users', TRUE),
    ('users.update', 'Update Users', 'users', TRUE),
    ('users.delete', 'Delete Users', 'users', TRUE),
    ('users.manage_roles', 'Manage User Roles', 'users', TRUE),
    -- Beneficiaries
    ('beneficiaries.create', 'Create Beneficiaries', 'beneficiaries', TRUE),
    ('beneficiaries.read', 'View Beneficiaries', 'beneficiaries', TRUE),
    ('beneficiaries.update', 'Update Beneficiaries', 'beneficiaries', TRUE),
    ('beneficiaries.delete', 'Delete Beneficiaries', 'beneficiaries', TRUE);

-- Assign permissions to admin (all permissions)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p WHERE r.name = 'admin';

-- Assign permissions to staff (CRUD on programs + beneficiaries, read users)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'staff' AND p.name IN (
    'programs.create', 'programs.read', 'programs.update', 'programs.delete',
    'beneficiaries.create', 'beneficiaries.read', 'beneficiaries.update', 'beneficiaries.delete',
    'users.read'
);

-- Assign permissions to viewer (read-only on everything)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'viewer' AND p.name IN (
    'programs.read', 'beneficiaries.read', 'users.read'
);

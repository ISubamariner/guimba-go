package pg

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// RoleRepoPG implements repository.RoleRepository using PostgreSQL.
type RoleRepoPG struct {
	pool *pgxpool.Pool
}

// NewRoleRepoPG creates a new PostgreSQL role repository.
func NewRoleRepoPG(pool *pgxpool.Pool) *RoleRepoPG {
	return &RoleRepoPG{pool: pool}
}

func (r *RoleRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	query := `SELECT id, name, display_name, description, is_system_role, is_active, created_at, updated_at
		FROM roles WHERE id = $1`

	role := &entity.Role{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.DisplayName, &role.Description,
		&role.IsSystemRole, &role.IsActive, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	perms, err := r.GetPermissionsByRoleID(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	role.Permissions = perms

	return role, nil
}

func (r *RoleRepoPG) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	query := `SELECT id, name, display_name, description, is_system_role, is_active, created_at, updated_at
		FROM roles WHERE name = $1`

	role := &entity.Role{}
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.DisplayName, &role.Description,
		&role.IsSystemRole, &role.IsActive, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	perms, err := r.GetPermissionsByRoleID(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	role.Permissions = perms

	return role, nil
}

func (r *RoleRepoPG) List(ctx context.Context) ([]*entity.Role, error) {
	query := `SELECT id, name, display_name, description, is_system_role, is_active, created_at, updated_at
		FROM roles ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*entity.Role
	for rows.Next() {
		role := &entity.Role{}
		if err := rows.Scan(
			&role.ID, &role.Name, &role.DisplayName, &role.Description,
			&role.IsSystemRole, &role.IsActive, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, err
		}

		perms, err := r.GetPermissionsByRoleID(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		role.Permissions = perms

		roles = append(roles, role)
	}

	return roles, rows.Err()
}

func (r *RoleRepoPG) GetPermissionsByRoleID(ctx context.Context, roleID uuid.UUID) ([]entity.Permission, error) {
	query := `
		SELECT p.id, p.name, p.display_name, p.description, p.category, p.is_system_permission, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.category, p.name`

	rows, err := r.pool.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []entity.Permission
	for rows.Next() {
		var p entity.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.DisplayName, &p.Description, &p.Category, &p.IsSystemPermission, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}

	return perms, rows.Err()
}

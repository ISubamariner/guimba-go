package pg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// UserRepoPG implements repository.UserRepository using PostgreSQL.
type UserRepoPG struct {
	pool *pgxpool.Pool
}

// NewUserRepoPG creates a new PostgreSQL user repository.
func NewUserRepoPG(pool *pgxpool.Pool) *UserRepoPG {
	return &UserRepoPG{pool: pool}
}

func (r *UserRepoPG) Create(ctx context.Context, u *entity.User) error {
	query := `
		INSERT INTO users (id, email, full_name, hashed_password, is_active, is_email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		u.ID, u.Email, u.FullName, u.HashedPassword, u.IsActive, u.IsEmailVerified, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (r *UserRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT u.id, u.email, u.full_name, u.hashed_password, u.is_active, u.is_email_verified,
		       u.last_login_at, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		WHERE u.id = $1 AND u.deleted_at IS NULL`

	u := &entity.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.FullName, &u.HashedPassword, &u.IsActive, &u.IsEmailVerified,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	roles, err := r.getUserRoles(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	u.Roles = roles

	return u, nil
}

func (r *UserRepoPG) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT u.id, u.email, u.full_name, u.hashed_password, u.is_active, u.is_email_verified,
		       u.last_login_at, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		WHERE u.email = $1 AND u.deleted_at IS NULL`

	u := &entity.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.FullName, &u.HashedPassword, &u.IsActive, &u.IsEmailVerified,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	roles, err := r.getUserRoles(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	u.Roles = roles

	return u, nil
}

func (r *UserRepoPG) List(ctx context.Context, filter repository.UserFilter) ([]*entity.User, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, "u.deleted_at IS NULL")

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("u.is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(u.full_name ILIKE $%d OR u.email ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	if filter.RoleName != nil && *filter.RoleName != "" {
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM user_roles ur JOIN roles ro ON ur.role_id = ro.id WHERE ur.user_id = u.id AND ro.name = $%d)", argIdx))
		args = append(args, *filter.RoleName)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	var total int
	countQuery := "SELECT COUNT(*) FROM users u " + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT u.id, u.email, u.full_name, u.hashed_password, u.is_active, u.is_email_verified,
		       u.last_login_at, u.created_at, u.updated_at, u.deleted_at
		FROM users u %s
		ORDER BY u.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		u := &entity.User{}
		if err := rows.Scan(
			&u.ID, &u.Email, &u.FullName, &u.HashedPassword, &u.IsActive, &u.IsEmailVerified,
			&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		roles, err := r.getUserRoles(ctx, u.ID)
		if err != nil {
			return nil, 0, err
		}
		u.Roles = roles
		users = append(users, u)
	}

	return users, total, rows.Err()
}

func (r *UserRepoPG) Update(ctx context.Context, u *entity.User) error {
	query := `
		UPDATE users
		SET full_name = $1, is_active = $2, is_email_verified = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL`

	_, err := r.pool.Exec(ctx, query, u.FullName, u.IsActive, u.IsEmailVerified, time.Now().UTC(), u.ID)
	return err
}

func (r *UserRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	return err
}

func (r *UserRepoPG) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, userID, roleID)
	return err
}

func (r *UserRepoPG) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.pool.Exec(ctx, query, userID, roleID)
	return err
}

func (r *UserRepoPG) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	return err
}

// getUserRoles retrieves all roles (with permissions) for a user.
func (r *UserRepoPG) getUserRoles(ctx context.Context, userID uuid.UUID) ([]entity.Role, error) {
	query := `
		SELECT r.id, r.name, r.display_name, r.description, r.is_system_role, r.is_active, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []entity.Role
	for rows.Next() {
		var role entity.Role
		if err := rows.Scan(
			&role.ID, &role.Name, &role.DisplayName, &role.Description,
			&role.IsSystemRole, &role.IsActive, &role.CreatedAt, &role.UpdatedAt,
		); err != nil {
			return nil, err
		}

		perms, err := r.getRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		role.Permissions = perms

		roles = append(roles, role)
	}

	return roles, rows.Err()
}

func (r *UserRepoPG) getRolePermissions(ctx context.Context, roleID uuid.UUID) ([]entity.Permission, error) {
	query := `
		SELECT p.id, p.name, p.display_name, p.description, p.category, p.is_system_permission, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1`

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

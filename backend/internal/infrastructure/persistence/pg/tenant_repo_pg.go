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

// TenantRepoPG implements repository.TenantRepository using PostgreSQL.
type TenantRepoPG struct {
	pool *pgxpool.Pool
}

// NewTenantRepoPG creates a new PostgreSQL tenant repository.
func NewTenantRepoPG(pool *pgxpool.Pool) *TenantRepoPG {
	return &TenantRepoPG{pool: pool}
}

func (r *TenantRepoPG) Create(ctx context.Context, t *entity.Tenant) error {
	query := `
		INSERT INTO tenants (id, full_name, email, phone_number, national_id,
			address_street, address_city, address_state_or_region, address_postal_code, address_country,
			landlord_id, user_id, is_active, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	var street, city, stateOrRegion, postalCode, country *string
	if t.Address != nil {
		street = &t.Address.Street
		city = &t.Address.City
		stateOrRegion = &t.Address.StateOrRegion
		if t.Address.PostalCode != "" {
			postalCode = &t.Address.PostalCode
		}
		country = &t.Address.Country
	}

	_, err := r.pool.Exec(ctx, query,
		t.ID, t.FullName, t.Email, t.PhoneNumber, t.NationalID,
		street, city, stateOrRegion, postalCode, country,
		t.LandlordID, t.UserID, t.IsActive, t.Notes, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (r *TenantRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	query := `
		SELECT id, full_name, email, phone_number, national_id,
			address_street, address_city, address_state_or_region, address_postal_code, address_country,
			landlord_id, user_id, is_active, notes, created_at, updated_at, deleted_at
		FROM tenants
		WHERE id = $1 AND deleted_at IS NULL`

	return r.scanTenant(r.pool.QueryRow(ctx, query, id))
}

func (r *TenantRepoPG) GetByEmail(ctx context.Context, email string) (*entity.Tenant, error) {
	query := `
		SELECT id, full_name, email, phone_number, national_id,
			address_street, address_city, address_state_or_region, address_postal_code, address_country,
			landlord_id, user_id, is_active, notes, created_at, updated_at, deleted_at
		FROM tenants
		WHERE email = $1 AND deleted_at IS NULL`

	return r.scanTenant(r.pool.QueryRow(ctx, query, email))
}

func (r *TenantRepoPG) List(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, "t.deleted_at IS NULL")

	if filter.LandlordID != nil {
		conditions = append(conditions, fmt.Sprintf("t.landlord_id = $%d", argIdx))
		args = append(args, *filter.LandlordID)
		argIdx++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("t.is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(t.full_name ILIKE $%d OR t.email ILIKE $%d OR t.phone_number ILIKE $%d)",
			argIdx, argIdx, argIdx,
		))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := "SELECT COUNT(*) FROM tenants t " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT t.id, t.full_name, t.email, t.phone_number, t.national_id,
			t.address_street, t.address_city, t.address_state_or_region, t.address_postal_code, t.address_country,
			t.landlord_id, t.user_id, t.is_active, t.notes, t.created_at, t.updated_at, t.deleted_at
		FROM tenants t %s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tenants []*entity.Tenant
	for rows.Next() {
		t, err := r.scanTenantRow(rows)
		if err != nil {
			return nil, 0, err
		}
		tenants = append(tenants, t)
	}

	return tenants, total, rows.Err()
}

func (r *TenantRepoPG) Update(ctx context.Context, t *entity.Tenant) error {
	query := `
		UPDATE tenants
		SET full_name = $1, email = $2, phone_number = $3, national_id = $4,
			address_street = $5, address_city = $6, address_state_or_region = $7,
			address_postal_code = $8, address_country = $9,
			is_active = $10, notes = $11, updated_at = $12
		WHERE id = $13 AND deleted_at IS NULL`

	var street, city, stateOrRegion, postalCode, country *string
	if t.Address != nil {
		street = &t.Address.Street
		city = &t.Address.City
		stateOrRegion = &t.Address.StateOrRegion
		if t.Address.PostalCode != "" {
			postalCode = &t.Address.PostalCode
		}
		country = &t.Address.Country
	}

	_, err := r.pool.Exec(ctx, query,
		t.FullName, t.Email, t.PhoneNumber, t.NationalID,
		street, city, stateOrRegion, postalCode, country,
		t.IsActive, t.Notes, time.Now().UTC(), t.ID,
	)
	return err
}

func (r *TenantRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tenants SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	return err
}

// scanTenant scans a single row into a Tenant entity.
func (r *TenantRepoPG) scanTenant(row pgx.Row) (*entity.Tenant, error) {
	t := &entity.Tenant{}
	var street, city, stateOrRegion, postalCode, country *string

	err := row.Scan(
		&t.ID, &t.FullName, &t.Email, &t.PhoneNumber, &t.NationalID,
		&street, &city, &stateOrRegion, &postalCode, &country,
		&t.LandlordID, &t.UserID, &t.IsActive, &t.Notes, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if street != nil || city != nil {
		pc := ""
		if postalCode != nil {
			pc = *postalCode
		}
		c := "Philippines"
		if country != nil {
			c = *country
		}
		s := ""
		if street != nil {
			s = *street
		}
		ci := ""
		if city != nil {
			ci = *city
		}
		sr := ""
		if stateOrRegion != nil {
			sr = *stateOrRegion
		}
		t.Address = &entity.Address{
			Street: s, City: ci, StateOrRegion: sr, PostalCode: pc, Country: c,
		}
	}

	return t, nil
}

// scanTenantRow scans a rows.Next() result into a Tenant entity.
func (r *TenantRepoPG) scanTenantRow(rows pgx.Rows) (*entity.Tenant, error) {
	t := &entity.Tenant{}
	var street, city, stateOrRegion, postalCode, country *string

	err := rows.Scan(
		&t.ID, &t.FullName, &t.Email, &t.PhoneNumber, &t.NationalID,
		&street, &city, &stateOrRegion, &postalCode, &country,
		&t.LandlordID, &t.UserID, &t.IsActive, &t.Notes, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	if street != nil || city != nil {
		pc := ""
		if postalCode != nil {
			pc = *postalCode
		}
		c := "Philippines"
		if country != nil {
			c = *country
		}
		s := ""
		if street != nil {
			s = *street
		}
		ci := ""
		if city != nil {
			ci = *city
		}
		sr := ""
		if stateOrRegion != nil {
			sr = *stateOrRegion
		}
		t.Address = &entity.Address{
			Street: s, City: ci, StateOrRegion: sr, PostalCode: pc, Country: c,
		}
	}

	return t, nil
}

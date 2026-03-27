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

// PropertyRepoPG implements repository.PropertyRepository using PostgreSQL.
type PropertyRepoPG struct {
	pool *pgxpool.Pool
}

// NewPropertyRepoPG creates a new PostgreSQL property repository.
func NewPropertyRepoPG(pool *pgxpool.Pool) *PropertyRepoPG {
	return &PropertyRepoPG{pool: pool}
}

func (r *PropertyRepoPG) Create(ctx context.Context, p *entity.Property) error {
	query := `
		INSERT INTO properties (id, name, property_code,
			address_street, address_city, address_state_or_region, address_postal_code, address_country,
			geojson_coordinates, property_type, size_in_acres, size_in_sqm,
			owner_id, is_available_for_rent, is_active, monthly_rent_amount, description,
			created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`

	var street, city, stateOrRegion, postalCode, country *string
	if p.Address != nil {
		street = &p.Address.Street
		city = &p.Address.City
		stateOrRegion = &p.Address.StateOrRegion
		if p.Address.PostalCode != "" {
			postalCode = &p.Address.PostalCode
		}
		country = &p.Address.Country
	}

	_, err := r.pool.Exec(ctx, query,
		p.ID, p.Name, p.PropertyCode,
		street, city, stateOrRegion, postalCode, country,
		p.GeoJSONCoordinates, p.PropertyType, p.SizeInAcres, p.SizeInSqm,
		p.OwnerID, p.IsAvailableForRent, p.IsActive, p.MonthlyRentAmount, p.Description,
		p.CreatedAt, p.UpdatedAt,
	)
	return err
}

// Column list used by GetByID, GetByPropertyCode, and List
const propertyColumns = `id, name, property_code,
	address_street, address_city, address_state_or_region, address_postal_code, address_country,
	geojson_coordinates, property_type, size_in_acres, size_in_sqm,
	owner_id, is_available_for_rent, is_active, monthly_rent_amount, description,
	created_at, updated_at, deleted_at`

func (r *PropertyRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
	query := `SELECT ` + propertyColumns + ` FROM properties WHERE id = $1 AND deleted_at IS NULL`
	return r.scanProperty(r.pool.QueryRow(ctx, query, id))
}

func (r *PropertyRepoPG) GetByPropertyCode(ctx context.Context, code string) (*entity.Property, error) {
	query := `SELECT ` + propertyColumns + ` FROM properties WHERE property_code = $1 AND deleted_at IS NULL`
	return r.scanProperty(r.pool.QueryRow(ctx, query, code))
}

func (r *PropertyRepoPG) List(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, "p.deleted_at IS NULL")

	if filter.OwnerID != nil {
		conditions = append(conditions, fmt.Sprintf("p.owner_id = $%d", argIdx))
		args = append(args, *filter.OwnerID)
		argIdx++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("p.is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}

	if filter.IsAvailableForRent != nil {
		conditions = append(conditions, fmt.Sprintf("p.is_available_for_rent = $%d", argIdx))
		args = append(args, *filter.IsAvailableForRent)
		argIdx++
	}

	if filter.PropertyType != nil && *filter.PropertyType != "" {
		conditions = append(conditions, fmt.Sprintf("p.property_type = $%d", argIdx))
		args = append(args, *filter.PropertyType)
		argIdx++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(p.name ILIKE $%d OR p.property_code ILIKE $%d)",
			argIdx, argIdx,
		))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := "SELECT COUNT(*) FROM properties p " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT p.id, p.name, p.property_code,
			p.address_street, p.address_city, p.address_state_or_region, p.address_postal_code, p.address_country,
			p.geojson_coordinates, p.property_type, p.size_in_acres, p.size_in_sqm,
			p.owner_id, p.is_available_for_rent, p.is_active, p.monthly_rent_amount, p.description,
			p.created_at, p.updated_at, p.deleted_at
		FROM properties p %s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var properties []*entity.Property
	for rows.Next() {
		p, err := r.scanPropertyRow(rows)
		if err != nil {
			return nil, 0, err
		}
		properties = append(properties, p)
	}

	return properties, total, rows.Err()
}

func (r *PropertyRepoPG) Update(ctx context.Context, p *entity.Property) error {
	query := `
		UPDATE properties
		SET name = $1, property_code = $2,
			address_street = $3, address_city = $4, address_state_or_region = $5,
			address_postal_code = $6, address_country = $7,
			geojson_coordinates = $8, property_type = $9, size_in_acres = $10, size_in_sqm = $11,
			is_available_for_rent = $12, is_active = $13, monthly_rent_amount = $14, description = $15,
			updated_at = $16
		WHERE id = $17 AND deleted_at IS NULL`

	var street, city, stateOrRegion, postalCode, country *string
	if p.Address != nil {
		street = &p.Address.Street
		city = &p.Address.City
		stateOrRegion = &p.Address.StateOrRegion
		if p.Address.PostalCode != "" {
			postalCode = &p.Address.PostalCode
		}
		country = &p.Address.Country
	}

	_, err := r.pool.Exec(ctx, query,
		p.Name, p.PropertyCode,
		street, city, stateOrRegion, postalCode, country,
		p.GeoJSONCoordinates, p.PropertyType, p.SizeInAcres, p.SizeInSqm,
		p.IsAvailableForRent, p.IsActive, p.MonthlyRentAmount, p.Description,
		time.Now().UTC(), p.ID,
	)
	return err
}

func (r *PropertyRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE properties SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	return err
}

// scanProperty scans a single row into a Property entity.
func (r *PropertyRepoPG) scanProperty(row pgx.Row) (*entity.Property, error) {
	p := &entity.Property{}
	var street, city, stateOrRegion, postalCode, country *string

	err := row.Scan(
		&p.ID, &p.Name, &p.PropertyCode,
		&street, &city, &stateOrRegion, &postalCode, &country,
		&p.GeoJSONCoordinates, &p.PropertyType, &p.SizeInAcres, &p.SizeInSqm,
		&p.OwnerID, &p.IsAvailableForRent, &p.IsActive, &p.MonthlyRentAmount, &p.Description,
		&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
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
		p.Address = &entity.Address{
			Street: s, City: ci, StateOrRegion: sr, PostalCode: pc, Country: c,
		}
	}

	return p, nil
}

// scanPropertyRow scans a rows.Next() result into a Property entity.
func (r *PropertyRepoPG) scanPropertyRow(rows pgx.Rows) (*entity.Property, error) {
	p := &entity.Property{}
	var street, city, stateOrRegion, postalCode, country *string

	err := rows.Scan(
		&p.ID, &p.Name, &p.PropertyCode,
		&street, &city, &stateOrRegion, &postalCode, &country,
		&p.GeoJSONCoordinates, &p.PropertyType, &p.SizeInAcres, &p.SizeInSqm,
		&p.OwnerID, &p.IsAvailableForRent, &p.IsActive, &p.MonthlyRentAmount, &p.Description,
		&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
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
		p.Address = &entity.Address{
			Street: s, City: ci, StateOrRegion: sr, PostalCode: pc, Country: c,
		}
	}

	return p, nil
}

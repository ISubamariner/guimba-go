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

// ProgramRepoPG implements repository.ProgramRepository using PostgreSQL.
type ProgramRepoPG struct {
	pool *pgxpool.Pool
}

// NewProgramRepoPG creates a new PostgreSQL program repository.
func NewProgramRepoPG(pool *pgxpool.Pool) *ProgramRepoPG {
	return &ProgramRepoPG{pool: pool}
}

func (r *ProgramRepoPG) Create(ctx context.Context, p *entity.Program) error {
	query := `
		INSERT INTO programs (id, name, description, status, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		p.ID, p.Name, p.Description, p.Status, p.StartDate, p.EndDate, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *ProgramRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*entity.Program, error) {
	query := `
		SELECT id, name, description, status, start_date, end_date, created_at, updated_at, deleted_at
		FROM programs
		WHERE id = $1 AND deleted_at IS NULL`

	p := &entity.Program{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Status,
		&p.StartDate, &p.EndDate, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func (r *ProgramRepoPG) List(ctx context.Context, filter repository.ProgramFilter) ([]*entity.Program, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM programs " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch paginated results
	dataQuery := fmt.Sprintf(`
		SELECT id, name, description, status, start_date, end_date, created_at, updated_at, deleted_at
		FROM programs %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var programs []*entity.Program
	for rows.Next() {
		p := &entity.Program{}
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Status,
			&p.StartDate, &p.EndDate, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		programs = append(programs, p)
	}

	return programs, total, rows.Err()
}

func (r *ProgramRepoPG) Update(ctx context.Context, p *entity.Program) error {
	query := `
		UPDATE programs
		SET name = $1, description = $2, status = $3, start_date = $4, end_date = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL`

	_, err := r.pool.Exec(ctx, query,
		p.Name, p.Description, p.Status, p.StartDate, p.EndDate, time.Now().UTC(), p.ID,
	)
	return err
}

func (r *ProgramRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE programs SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	return err
}

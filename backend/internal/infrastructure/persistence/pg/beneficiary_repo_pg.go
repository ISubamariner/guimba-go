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

// BeneficiaryRepoPG implements repository.BeneficiaryRepository using PostgreSQL.
type BeneficiaryRepoPG struct {
	pool *pgxpool.Pool
}

// NewBeneficiaryRepoPG creates a new PostgreSQL beneficiary repository.
func NewBeneficiaryRepoPG(pool *pgxpool.Pool) *BeneficiaryRepoPG {
	return &BeneficiaryRepoPG{pool: pool}
}

func (r *BeneficiaryRepoPG) Create(ctx context.Context, b *entity.Beneficiary) error {
	query := `
		INSERT INTO beneficiaries (id, full_name, email, phone_number, national_id, address, date_of_birth, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.pool.Exec(ctx, query,
		b.ID, b.FullName, b.Email, b.PhoneNumber, b.NationalID, b.Address, b.DateOfBirth, b.Status, b.Notes, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *BeneficiaryRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error) {
	query := `
		SELECT id, full_name, email, phone_number, national_id, address, date_of_birth, status, notes, created_at, updated_at, deleted_at
		FROM beneficiaries
		WHERE id = $1 AND deleted_at IS NULL`

	b := &entity.Beneficiary{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.FullName, &b.Email, &b.PhoneNumber, &b.NationalID, &b.Address, &b.DateOfBirth, &b.Status, &b.Notes, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Load program enrollments
	enrollments, err := r.loadEnrollments(ctx, id)
	if err != nil {
		return nil, err
	}
	b.Programs = enrollments

	return b, nil
}

func (r *BeneficiaryRepoPG) List(ctx context.Context, filter repository.BeneficiaryFilter) ([]*entity.Beneficiary, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, "b.deleted_at IS NULL")

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("b.status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}

	if filter.ProgramID != nil {
		conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM program_beneficiaries pb WHERE pb.beneficiary_id = b.id AND pb.program_id = $%d)", argIdx))
		args = append(args, *filter.ProgramID)
		argIdx++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("b.full_name ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM beneficiaries b " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch paginated results
	dataQuery := fmt.Sprintf(`
		SELECT b.id, b.full_name, b.email, b.phone_number, b.national_id, b.address, b.date_of_birth, b.status, b.notes, b.created_at, b.updated_at, b.deleted_at
		FROM beneficiaries b %s
		ORDER BY b.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var beneficiaries []*entity.Beneficiary
	for rows.Next() {
		b := &entity.Beneficiary{}
		if err := rows.Scan(
			&b.ID, &b.FullName, &b.Email, &b.PhoneNumber, &b.NationalID, &b.Address, &b.DateOfBirth, &b.Status, &b.Notes, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		beneficiaries = append(beneficiaries, b)
	}

	return beneficiaries, total, rows.Err()
}

func (r *BeneficiaryRepoPG) Update(ctx context.Context, b *entity.Beneficiary) error {
	query := `
		UPDATE beneficiaries
		SET full_name = $1, email = $2, phone_number = $3, national_id = $4, address = $5, date_of_birth = $6, status = $7, notes = $8, updated_at = $9
		WHERE id = $10 AND deleted_at IS NULL`

	_, err := r.pool.Exec(ctx, query,
		b.FullName, b.Email, b.PhoneNumber, b.NationalID, b.Address, b.DateOfBirth, b.Status, b.Notes, time.Now().UTC(), b.ID,
	)
	return err
}

func (r *BeneficiaryRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE beneficiaries SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	return err
}

func (r *BeneficiaryRepoPG) EnrollInProgram(ctx context.Context, beneficiaryID, programID uuid.UUID) error {
	query := `
		INSERT INTO program_beneficiaries (program_id, beneficiary_id)
		VALUES ($1, $2)
		ON CONFLICT (program_id, beneficiary_id) DO NOTHING`

	result, err := r.pool.Exec(ctx, query, programID, beneficiaryID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return entity.ErrBeneficiaryAlreadyEnrolled
	}
	return nil
}

func (r *BeneficiaryRepoPG) RemoveFromProgram(ctx context.Context, beneficiaryID, programID uuid.UUID) error {
	query := `DELETE FROM program_beneficiaries WHERE beneficiary_id = $1 AND program_id = $2`
	result, err := r.pool.Exec(ctx, query, beneficiaryID, programID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return entity.ErrBeneficiaryNotEnrolled
	}
	return nil
}

func (r *BeneficiaryRepoPG) loadEnrollments(ctx context.Context, beneficiaryID uuid.UUID) ([]entity.ProgramEnrollment, error) {
	query := `
		SELECT pb.program_id, p.name, pb.enrolled_at, pb.status
		FROM program_beneficiaries pb
		JOIN programs p ON p.id = pb.program_id
		WHERE pb.beneficiary_id = $1
		ORDER BY pb.enrolled_at DESC`

	rows, err := r.pool.Query(ctx, query, beneficiaryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []entity.ProgramEnrollment
	for rows.Next() {
		var e entity.ProgramEnrollment
		if err := rows.Scan(&e.ProgramID, &e.ProgramName, &e.EnrolledAt, &e.Status); err != nil {
			return nil, err
		}
		enrollments = append(enrollments, e)
	}

	return enrollments, rows.Err()
}

package pg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type DebtRepoPG struct {
	pool *pgxpool.Pool
}

func NewDebtRepoPG(pool *pgxpool.Pool) *DebtRepoPG {
	return &DebtRepoPG{pool: pool}
}

const debtColumns = `id, tenant_id, landlord_id, property_id, debt_type, description,
	original_amount, original_currency, amount_paid, amount_paid_currency,
	due_date, status, notes, created_at, updated_at, deleted_at`

func (r *DebtRepoPG) Create(ctx context.Context, d *entity.Debt) error {
	query := `
		INSERT INTO debts (id, tenant_id, landlord_id, property_id, debt_type, description,
			original_amount, original_currency, amount_paid, amount_paid_currency,
			due_date, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.pool.Exec(ctx, query,
		d.ID, d.TenantID, d.LandlordID, d.PropertyID, string(d.DebtType), d.Description,
		d.OriginalAmount.Amount, string(d.OriginalAmount.Currency),
		d.AmountPaid.Amount, string(d.AmountPaid.Currency),
		d.DueDate, string(d.Status), d.Notes, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (r *DebtRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
	query := `SELECT ` + debtColumns + ` FROM debts WHERE id = $1 AND deleted_at IS NULL`
	return r.scanDebt(r.pool.QueryRow(ctx, query, id))
}

func (r *DebtRepoPG) List(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, "d.deleted_at IS NULL")

	if filter.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("d.tenant_id = $%d", argIdx))
		args = append(args, *filter.TenantID)
		argIdx++
	}
	if filter.LandlordID != nil {
		conditions = append(conditions, fmt.Sprintf("d.landlord_id = $%d", argIdx))
		args = append(args, *filter.LandlordID)
		argIdx++
	}
	if filter.PropertyID != nil {
		conditions = append(conditions, fmt.Sprintf("d.property_id = $%d", argIdx))
		args = append(args, *filter.PropertyID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("d.status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.DebtType != nil {
		conditions = append(conditions, fmt.Sprintf("d.debt_type = $%d", argIdx))
		args = append(args, string(*filter.DebtType))
		argIdx++
	}
	if filter.IsOverdue != nil && *filter.IsOverdue {
		conditions = append(conditions, "d.due_date < NOW()")
		conditions = append(conditions, "d.status NOT IN ('PAID', 'CANCELLED')")
	}
	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("d.description ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := "SELECT COUNT(*) FROM debts d " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT d.id, d.tenant_id, d.landlord_id, d.property_id, d.debt_type, d.description,
			d.original_amount, d.original_currency, d.amount_paid, d.amount_paid_currency,
			d.due_date, d.status, d.notes, d.created_at, d.updated_at, d.deleted_at
		FROM debts d %s
		ORDER BY d.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var debts []*entity.Debt
	for rows.Next() {
		d, err := r.scanDebtRow(rows)
		if err != nil {
			return nil, 0, err
		}
		debts = append(debts, d)
	}

	return debts, total, rows.Err()
}

func (r *DebtRepoPG) Update(ctx context.Context, d *entity.Debt) error {
	query := `
		UPDATE debts
		SET debt_type = $1, description = $2,
			original_amount = $3, original_currency = $4,
			amount_paid = $5, amount_paid_currency = $6,
			due_date = $7, status = $8, notes = $9,
			property_id = $10
		WHERE id = $11 AND deleted_at IS NULL`

	_, err := r.pool.Exec(ctx, query,
		string(d.DebtType), d.Description,
		d.OriginalAmount.Amount, string(d.OriginalAmount.Currency),
		d.AmountPaid.Amount, string(d.AmountPaid.Currency),
		d.DueDate, string(d.Status), d.Notes,
		d.PropertyID, d.ID,
	)
	return err
}

func (r *DebtRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE debts SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	return err
}

func (r *DebtRepoPG) HasActiveDebtsForProperty(ctx context.Context, propertyID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM debts WHERE property_id = $1 AND status NOT IN ('PAID', 'CANCELLED') AND deleted_at IS NULL)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, propertyID).Scan(&exists)
	return exists, err
}

func (r *DebtRepoPG) scanDebt(row pgx.Row) (*entity.Debt, error) {
	d := &entity.Debt{}
	var origAmount decimal.Decimal
	var origCurrency string
	var paidAmount decimal.Decimal
	var paidCurrency string
	var debtType, status string

	err := row.Scan(
		&d.ID, &d.TenantID, &d.LandlordID, &d.PropertyID, &debtType, &d.Description,
		&origAmount, &origCurrency, &paidAmount, &paidCurrency,
		&d.DueDate, &status, &d.Notes, &d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	d.DebtType = entity.DebtType(debtType)
	d.Status = entity.DebtStatus(status)
	d.OriginalAmount = entity.Money{Amount: origAmount, Currency: entity.Currency(origCurrency)}
	d.AmountPaid = entity.Money{Amount: paidAmount, Currency: entity.Currency(paidCurrency)}

	return d, nil
}

func (r *DebtRepoPG) scanDebtRow(rows pgx.Rows) (*entity.Debt, error) {
	d := &entity.Debt{}
	var origAmount decimal.Decimal
	var origCurrency string
	var paidAmount decimal.Decimal
	var paidCurrency string
	var debtType, status string

	err := rows.Scan(
		&d.ID, &d.TenantID, &d.LandlordID, &d.PropertyID, &debtType, &d.Description,
		&origAmount, &origCurrency, &paidAmount, &paidCurrency,
		&d.DueDate, &status, &d.Notes, &d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	d.DebtType = entity.DebtType(debtType)
	d.Status = entity.DebtStatus(status)
	d.OriginalAmount = entity.Money{Amount: origAmount, Currency: entity.Currency(origCurrency)}
	d.AmountPaid = entity.Money{Amount: paidAmount, Currency: entity.Currency(paidCurrency)}

	return d, nil
}

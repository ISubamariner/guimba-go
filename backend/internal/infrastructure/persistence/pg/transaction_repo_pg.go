package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type TransactionRepoPG struct {
	pool *pgxpool.Pool
}

func NewTransactionRepoPG(pool *pgxpool.Pool) *TransactionRepoPG {
	return &TransactionRepoPG{pool: pool}
}

const transactionColumns = `id, debt_id, tenant_id, landlord_id, recorded_by_user_id,
	transaction_type, amount, currency, payment_method, transaction_date,
	description, receipt_number, reference_number,
	is_verified, verified_by_user_id, verified_at, created_at, updated_at`

func (r *TransactionRepoPG) Create(ctx context.Context, tx *entity.Transaction) error {
	query := `
		INSERT INTO transactions (id, debt_id, tenant_id, landlord_id, recorded_by_user_id,
			transaction_type, amount, currency, payment_method, transaction_date,
			description, receipt_number, reference_number,
			is_verified, verified_by_user_id, verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	_, err := r.pool.Exec(ctx, query,
		tx.ID, tx.DebtID, tx.TenantID, tx.LandlordID, tx.RecordedByUserID,
		string(tx.TransactionType), tx.Amount.Amount, string(tx.Amount.Currency),
		string(tx.PaymentMethod), tx.TransactionDate,
		tx.Description, tx.ReceiptNumber, tx.ReferenceNumber,
		tx.IsVerified, tx.VerifiedByUserID, tx.VerifiedAt,
		tx.CreatedAt, tx.UpdatedAt,
	)
	return err
}

func (r *TransactionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	query := `SELECT ` + transactionColumns + ` FROM transactions WHERE id = $1`
	return r.scanTransaction(r.pool.QueryRow(ctx, query, id))
}

func (r *TransactionRepoPG) List(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.DebtID != nil {
		conditions = append(conditions, fmt.Sprintf("t.debt_id = $%d", argIdx))
		args = append(args, *filter.DebtID)
		argIdx++
	}
	if filter.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("t.tenant_id = $%d", argIdx))
		args = append(args, *filter.TenantID)
		argIdx++
	}
	if filter.LandlordID != nil {
		conditions = append(conditions, fmt.Sprintf("t.landlord_id = $%d", argIdx))
		args = append(args, *filter.LandlordID)
		argIdx++
	}
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("t.transaction_type = $%d", argIdx))
		args = append(args, string(*filter.Type))
		argIdx++
	}
	if filter.IsVerified != nil {
		conditions = append(conditions, fmt.Sprintf("t.is_verified = $%d", argIdx))
		args = append(args, *filter.IsVerified)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM transactions t " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT t.id, t.debt_id, t.tenant_id, t.landlord_id, t.recorded_by_user_id,
			t.transaction_type, t.amount, t.currency, t.payment_method, t.transaction_date,
			t.description, t.receipt_number, t.reference_number,
			t.is_verified, t.verified_by_user_id, t.verified_at, t.created_at, t.updated_at
		FROM transactions t %s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txs []*entity.Transaction
	for rows.Next() {
		tx, err := r.scanTransactionRow(rows)
		if err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}

	return txs, total, rows.Err()
}

func (r *TransactionRepoPG) Update(ctx context.Context, tx *entity.Transaction) error {
	query := `
		UPDATE transactions
		SET is_verified = $1, verified_by_user_id = $2, verified_at = $3
		WHERE id = $4`

	_, err := r.pool.Exec(ctx, query,
		tx.IsVerified, tx.VerifiedByUserID, tx.VerifiedAt, tx.ID,
	)
	return err
}

func (r *TransactionRepoPG) ExistsByReferenceNumber(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM transactions WHERE debt_id = $1 AND reference_number = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, debtID, refNum).Scan(&exists)
	return exists, err
}

func (r *TransactionRepoPG) scanTransaction(row pgx.Row) (*entity.Transaction, error) {
	tx := &entity.Transaction{}
	var txType, currency, method string
	var amount decimal.Decimal

	err := row.Scan(
		&tx.ID, &tx.DebtID, &tx.TenantID, &tx.LandlordID, &tx.RecordedByUserID,
		&txType, &amount, &currency, &method, &tx.TransactionDate,
		&tx.Description, &tx.ReceiptNumber, &tx.ReferenceNumber,
		&tx.IsVerified, &tx.VerifiedByUserID, &tx.VerifiedAt,
		&tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	tx.TransactionType = entity.TransactionType(txType)
	tx.PaymentMethod = entity.PaymentMethod(method)
	tx.Amount = entity.Money{Amount: amount, Currency: entity.Currency(currency)}

	return tx, nil
}

func (r *TransactionRepoPG) scanTransactionRow(rows pgx.Rows) (*entity.Transaction, error) {
	tx := &entity.Transaction{}
	var txType, currency, method string
	var amount decimal.Decimal

	err := rows.Scan(
		&tx.ID, &tx.DebtID, &tx.TenantID, &tx.LandlordID, &tx.RecordedByUserID,
		&txType, &amount, &currency, &method, &tx.TransactionDate,
		&tx.Description, &tx.ReceiptNumber, &tx.ReferenceNumber,
		&tx.IsVerified, &tx.VerifiedByUserID, &tx.VerifiedAt,
		&tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	tx.TransactionType = entity.TransactionType(txType)
	tx.PaymentMethod = entity.PaymentMethod(method)
	tx.Amount = entity.Money{Amount: amount, Currency: entity.Currency(currency)}

	return tx, nil
}

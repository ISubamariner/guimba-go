package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	txuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/transaction"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/validator"
)

type TransactionHandler struct {
	recordPaymentUC *txuc.RecordPaymentUseCase
	recordRefundUC  *txuc.RecordRefundUseCase
	getUC           *txuc.GetTransactionUseCase
	listUC          *txuc.ListTransactionsUseCase
	verifyUC        *txuc.VerifyTransactionUseCase
}

func NewTransactionHandler(
	recordPaymentUC *txuc.RecordPaymentUseCase,
	recordRefundUC *txuc.RecordRefundUseCase,
	getUC *txuc.GetTransactionUseCase,
	listUC *txuc.ListTransactionsUseCase,
	verifyUC *txuc.VerifyTransactionUseCase,
) *TransactionHandler {
	return &TransactionHandler{
		recordPaymentUC: recordPaymentUC,
		recordRefundUC:  recordRefundUC,
		getUC:           getUC,
		listUC:          listUC,
		verifyUC:        verifyUC,
	}
}

// RecordPayment godoc
// @Summary      Record a payment
// @Description  Records a payment transaction against a debt
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RecordPaymentRequest  true  "Payment data"
// @Success      201   {object}  dto.TransactionResponse
// @Failure      400,404,409,422  {object}  apperror.ErrorResponse
// @Router       /api/v1/transactions/payment [post]
func (h *TransactionHandler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	var req dto.RecordPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}
	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	recorderID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))

	amount, err := dto.ToMoney(req.Amount)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	txDate, err := time.Parse(time.RFC3339, req.TransactionDate)
	if err != nil {
		txDate, err = time.Parse("2006-01-02", req.TransactionDate)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid transaction_date format"))
			return
		}
	}

	tx, err := h.recordPaymentUC.Execute(r.Context(), req.DebtID, req.TenantID, &recorderID, amount, entity.PaymentMethod(req.PaymentMethod), txDate, req.Description, req.ReceiptNumber, req.ReferenceNumber)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewTransactionResponse(tx))
}

// RecordRefund godoc
// @Summary      Record a refund
// @Description  Records a refund transaction against a debt
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RecordRefundRequest  true  "Refund data"
// @Success      201   {object}  dto.TransactionResponse
// @Failure      400,404,409,422  {object}  apperror.ErrorResponse
// @Router       /api/v1/transactions/refund [post]
func (h *TransactionHandler) RecordRefund(w http.ResponseWriter, r *http.Request) {
	var req dto.RecordRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}
	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	recorderID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))

	amount, err := dto.ToMoney(req.Amount)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	refundDate, err := time.Parse(time.RFC3339, req.RefundDate)
	if err != nil {
		refundDate, err = time.Parse("2006-01-02", req.RefundDate)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid refund_date format"))
			return
		}
	}

	tx, err := h.recordRefundUC.Execute(r.Context(), req.DebtID, req.TenantID, &recorderID, amount, entity.PaymentMethod(req.PaymentMethod), refundDate, req.Description, req.ReferenceNumber)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewTransactionResponse(tx))
}

// Get godoc
// @Summary      Get a transaction
// @Description  Retrieves a transaction by ID
// @Tags         transactions
// @Produce      json
// @Param        id   path      string  true  "Transaction ID (UUID)"
// @Success      200  {object}  dto.TransactionResponse
// @Failure      400,404  {object}  apperror.ErrorResponse
// @Router       /api/v1/transactions/{id} [get]
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid transaction ID format"))
		return
	}

	tx, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTransactionResponse(tx))
}

// List godoc
// @Summary      List transactions
// @Description  Returns a paginated list of transactions with optional filtering
// @Tags         transactions
// @Produce      json
// @Param        debt_id     query  string  false  "Filter by debt ID"
// @Param        tenant_id   query  string  false  "Filter by tenant ID"
// @Param        type        query  string  false  "Filter by transaction type"
// @Param        is_verified query  bool    false  "Filter by verification status"
// @Param        limit       query  int     false  "Page size (default 20, max 100)"
// @Param        offset      query  int     false  "Offset"
// @Success      200         {object}  dto.TransactionListResponse
// @Router       /api/v1/transactions [get]
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	landlordID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))
	filter := repository.TransactionFilter{
		LandlordID: &landlordID,
		Limit:      20,
		Offset:     0,
	}

	if s := r.URL.Query().Get("debt_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid debt_id filter"))
			return
		}
		filter.DebtID = &id
	}
	if s := r.URL.Query().Get("tenant_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid tenant_id filter"))
			return
		}
		filter.TenantID = &id
	}
	if s := r.URL.Query().Get("type"); s != "" {
		txType := entity.TransactionType(s)
		filter.Type = &txType
	}
	if s := r.URL.Query().Get("is_verified"); s != "" {
		v := s == "true"
		filter.IsVerified = &v
	}
	if s := r.URL.Query().Get("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			filter.Limit = v
		}
	}
	if s := r.URL.Query().Get("offset"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			filter.Offset = v
		}
	}

	txs, total, err := h.listUC.Execute(r.Context(), filter)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTransactionListResponse(txs, total, filter.Limit, filter.Offset))
}

// Verify godoc
// @Summary      Verify a transaction
// @Description  Marks a transaction as verified by the authenticated user
// @Tags         transactions
// @Produce      json
// @Param        id   path      string  true  "Transaction ID (UUID)"
// @Success      200  {object}  dto.TransactionResponse
// @Failure      400,404,409  {object}  apperror.ErrorResponse
// @Router       /api/v1/transactions/{id}/verify [put]
func (h *TransactionHandler) Verify(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid transaction ID format"))
		return
	}

	verifierID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))

	if err := h.verifyUC.Execute(r.Context(), id, verifierID); err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	tx, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTransactionResponse(tx))
}

func handleTransactionDomainError(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		apperror.WriteError(w, appErr)
		return
	}

	switch {
	case errors.Is(err, entity.ErrTransactionAmountRequired),
		errors.Is(err, entity.ErrTransactionInvalidType),
		errors.Is(err, entity.ErrTransactionInvalidPaymentMethod),
		errors.Is(err, entity.ErrTransactionDateRequired),
		errors.Is(err, entity.ErrCurrencyMismatch),
		errors.Is(err, entity.ErrNegativeAmount),
		errors.Is(err, entity.ErrInvalidCurrency),
		errors.Is(err, entity.ErrDebtDescriptionRequired):
		apperror.WriteError(w, apperror.NewValidation(err.Error()))
	case errors.Is(err, entity.ErrTransactionAlreadyVerified),
		errors.Is(err, entity.ErrTransactionDuplicateReference),
		errors.Is(err, entity.ErrDebtAlreadyPaid),
		errors.Is(err, entity.ErrDebtAlreadyCancelled),
		errors.Is(err, entity.ErrDebtOverpayment),
		errors.Is(err, entity.ErrInsufficientAmount):
		apperror.WriteError(w, apperror.NewConflict(err.Error()))
	default:
		slog.Error("unhandled error in transaction handler", "error", err)
		apperror.WriteError(w, apperror.NewInternal(err))
	}
}

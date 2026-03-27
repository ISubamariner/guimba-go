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
	debtuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/debt"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/validator"
)

type DebtHandler struct {
	createUC   *debtuc.CreateDebtUseCase
	getUC      *debtuc.GetDebtUseCase
	listUC     *debtuc.ListDebtsUseCase
	updateUC   *debtuc.UpdateDebtUseCase
	cancelUC   *debtuc.CancelDebtUseCase
	markPaidUC *debtuc.MarkDebtPaidUseCase
	deleteUC   *debtuc.DeleteDebtUseCase
}

func NewDebtHandler(
	createUC *debtuc.CreateDebtUseCase,
	getUC *debtuc.GetDebtUseCase,
	listUC *debtuc.ListDebtsUseCase,
	updateUC *debtuc.UpdateDebtUseCase,
	cancelUC *debtuc.CancelDebtUseCase,
	markPaidUC *debtuc.MarkDebtPaidUseCase,
	deleteUC *debtuc.DeleteDebtUseCase,
) *DebtHandler {
	return &DebtHandler{
		createUC: createUC, getUC: getUC, listUC: listUC,
		updateUC: updateUC, cancelUC: cancelUC, markPaidUC: markPaidUC,
		deleteUC: deleteUC,
	}
}

// Create godoc
// @Summary      Create a debt
// @Description  Creates a new debt for a tenant
// @Tags         debts
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateDebtRequest  true  "Debt data"
// @Success      201   {object}  dto.DebtResponse
// @Failure      400,404,422  {object}  apperror.ErrorResponse
// @Router       /api/v1/debts [post]
func (h *DebtHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateDebtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}
	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	landlordID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))

	amount, err := dto.ToMoney(req.OriginalAmount)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		dueDate, err = time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid due_date format (use RFC3339 or YYYY-MM-DD)"))
			return
		}
	}

	d, err := entity.NewDebt(req.TenantID, landlordID, req.PropertyID, entity.DebtType(req.DebtType), req.Description, amount, dueDate, req.Notes)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	if err := h.createUC.Execute(r.Context(), d); err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewDebtResponse(d))
}

// Get godoc
// @Summary      Get a debt
// @Description  Retrieves a debt by ID
// @Tags         debts
// @Produce      json
// @Param        id   path      string  true  "Debt ID (UUID)"
// @Success      200  {object}  dto.DebtResponse
// @Failure      400,404  {object}  apperror.ErrorResponse
// @Router       /api/v1/debts/{id} [get]
func (h *DebtHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid debt ID format"))
		return
	}

	d, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewDebtResponse(d))
}

// List godoc
// @Summary      List debts
// @Description  Returns a paginated list of debts with optional filtering
// @Tags         debts
// @Produce      json
// @Param        tenant_id   query  string  false  "Filter by tenant ID"
// @Param        property_id query  string  false  "Filter by property ID"
// @Param        status      query  string  false  "Filter by status"
// @Param        debt_type   query  string  false  "Filter by debt type"
// @Param        search      query  string  false  "Search description"
// @Param        limit       query  int     false  "Page size (default 20, max 100)"
// @Param        offset      query  int     false  "Offset"
// @Success      200         {object}  dto.DebtListResponse
// @Router       /api/v1/debts [get]
func (h *DebtHandler) List(w http.ResponseWriter, r *http.Request) {
	landlordID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))
	filter := repository.DebtFilter{
		LandlordID: &landlordID,
		Limit:      20,
		Offset:     0,
	}

	if s := r.URL.Query().Get("tenant_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid tenant_id filter"))
			return
		}
		filter.TenantID = &id
	}
	if s := r.URL.Query().Get("property_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid property_id filter"))
			return
		}
		filter.PropertyID = &id
	}
	if s := r.URL.Query().Get("status"); s != "" {
		status := entity.DebtStatus(s)
		if !status.IsValid() {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid status filter"))
			return
		}
		filter.Status = &status
	}
	if s := r.URL.Query().Get("debt_type"); s != "" {
		dt := entity.DebtType(s)
		if !dt.IsValid() {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid debt_type filter"))
			return
		}
		filter.DebtType = &dt
	}
	if s := r.URL.Query().Get("search"); s != "" {
		filter.Search = &s
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

	debts, total, err := h.listUC.Execute(r.Context(), filter)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewDebtListResponse(debts, total, filter.Limit, filter.Offset))
}

// Update godoc
// @Summary      Update a debt
// @Description  Updates mutable fields of an existing debt
// @Tags         debts
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "Debt ID (UUID)"
// @Param        body  body      dto.UpdateDebtRequest  true  "Updated debt data"
// @Success      200   {object}  dto.DebtResponse
// @Failure      400,404,422  {object}  apperror.ErrorResponse
// @Router       /api/v1/debts/{id} [put]
func (h *DebtHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid debt ID format"))
		return
	}

	var req dto.UpdateDebtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}
	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		dueDate, err = time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid due_date format"))
			return
		}
	}

	updates := &entity.Debt{
		Description: req.Description,
		DebtType:    entity.DebtType(req.DebtType),
		DueDate:     dueDate,
		PropertyID:  req.PropertyID,
		Notes:       req.Notes,
	}

	if err := h.updateUC.Execute(r.Context(), id, updates); err != nil {
		handleDebtDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewDebtResponse(updated))
}

// Cancel godoc
// @Summary      Cancel a debt
// @Description  Cancels a debt with an optional reason
// @Tags         debts
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "Debt ID (UUID)"
// @Param        body  body      dto.CancelDebtRequest  false "Cancel reason"
// @Success      200   {object}  dto.DebtResponse
// @Failure      400,404,409  {object}  apperror.ErrorResponse
// @Router       /api/v1/debts/{id}/cancel [put]
func (h *DebtHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid debt ID format"))
		return
	}

	var req dto.CancelDebtRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // optional body

	if err := h.cancelUC.Execute(r.Context(), id, req.Reason); err != nil {
		handleDebtDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewDebtResponse(updated))
}

// MarkPaid godoc
// @Summary      Mark debt as paid
// @Description  Forces a debt to PAID status by paying the remaining balance
// @Tags         debts
// @Produce      json
// @Param        id   path      string  true  "Debt ID (UUID)"
// @Success      200  {object}  dto.DebtResponse
// @Failure      400,404,409  {object}  apperror.ErrorResponse
// @Router       /api/v1/debts/{id}/pay [put]
func (h *DebtHandler) MarkPaid(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid debt ID format"))
		return
	}

	if err := h.markPaidUC.Execute(r.Context(), id); err != nil {
		handleDebtDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewDebtResponse(updated))
}

// Delete godoc
// @Summary      Delete a debt
// @Description  Soft-deletes a debt by ID
// @Tags         debts
// @Param        id   path      string  true  "Debt ID (UUID)"
// @Success      204
// @Failure      400,404  {object}  apperror.ErrorResponse
// @Router       /api/v1/debts/{id} [delete]
func (h *DebtHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid debt ID format"))
		return
	}

	if err := h.deleteUC.Execute(r.Context(), id); err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleDebtDomainError(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		apperror.WriteError(w, appErr)
		return
	}

	switch {
	case errors.Is(err, entity.ErrDebtDescriptionRequired),
		errors.Is(err, entity.ErrDebtDescriptionTooLong),
		errors.Is(err, entity.ErrDebtAmountRequired),
		errors.Is(err, entity.ErrDebtInvalidType),
		errors.Is(err, entity.ErrDebtDueDateRequired),
		errors.Is(err, entity.ErrCurrencyMismatch),
		errors.Is(err, entity.ErrNegativeAmount),
		errors.Is(err, entity.ErrInvalidCurrency):
		apperror.WriteError(w, apperror.NewValidation(err.Error()))
	case errors.Is(err, entity.ErrDebtAlreadyPaid),
		errors.Is(err, entity.ErrDebtAlreadyCancelled),
		errors.Is(err, entity.ErrDebtOverpayment),
		errors.Is(err, entity.ErrDebtNotPayable):
		apperror.WriteError(w, apperror.NewConflict(err.Error()))
	default:
		slog.Error("unhandled error in debt handler", "error", err)
		apperror.WriteError(w, apperror.NewInternal(err))
	}
}

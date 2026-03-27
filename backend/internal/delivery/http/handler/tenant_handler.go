package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	tenantuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/tenant"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/validator"
)

// TenantHandler handles HTTP requests for tenants.
type TenantHandler struct {
	createUC     *tenantuc.CreateTenantUseCase
	getUC        *tenantuc.GetTenantUseCase
	listUC       *tenantuc.ListTenantsUseCase
	updateUC     *tenantuc.UpdateTenantUseCase
	deactivateUC *tenantuc.DeactivateTenantUseCase
	deleteUC     *tenantuc.DeleteTenantUseCase
}

// NewTenantHandler creates a new TenantHandler.
func NewTenantHandler(
	createUC *tenantuc.CreateTenantUseCase,
	getUC *tenantuc.GetTenantUseCase,
	listUC *tenantuc.ListTenantsUseCase,
	updateUC *tenantuc.UpdateTenantUseCase,
	deactivateUC *tenantuc.DeactivateTenantUseCase,
	deleteUC *tenantuc.DeleteTenantUseCase,
) *TenantHandler {
	return &TenantHandler{
		createUC:     createUC,
		getUC:        getUC,
		listUC:       listUC,
		updateUC:     updateUC,
		deactivateUC: deactivateUC,
		deleteUC:     deleteUC,
	}
}

// Create godoc
// @Summary      Create a tenant
// @Description  Creates a new tenant belonging to the authenticated landlord
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateTenantRequest  true  "Tenant data"
// @Success      201   {object}  dto.TenantResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      409   {object}  apperror.ErrorResponse
// @Failure      422   {object}  apperror.ErrorResponse
// @Failure      500   {object}  apperror.ErrorResponse
// @Router       /api/v1/tenants [post]
func (h *TenantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	landlordID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))

	tenant, err := req.ToEntity(landlordID)
	if err != nil {
		handleTenantDomainError(w, err)
		return
	}

	if err := h.createUC.Execute(r.Context(), tenant); err != nil {
		handleTenantDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewTenantResponse(tenant))
}

// Get godoc
// @Summary      Get a tenant
// @Description  Retrieves a tenant by ID
// @Tags         tenants
// @Produce      json
// @Param        id   path      string  true  "Tenant ID (UUID)"
// @Success      200  {object}  dto.TenantResponse
// @Failure      400  {object}  apperror.ErrorResponse
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/tenants/{id} [get]
func (h *TenantHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid tenant ID format"))
		return
	}

	tenant, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleTenantDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTenantResponse(tenant))
}

// List godoc
// @Summary      List tenants
// @Description  Returns a paginated list of tenants with optional filtering
// @Tags         tenants
// @Produce      json
// @Param        landlord_id  query     string  false  "Filter by landlord ID"
// @Param        is_active    query     bool    false  "Filter by active status"
// @Param        search       query     string  false  "Search by name, email, phone"
// @Param        limit        query     int     false  "Page size (default 20, max 100)"
// @Param        offset       query     int     false  "Offset (default 0)"
// @Success      200          {object}  dto.TenantListResponse
// @Failure      500          {object}  apperror.ErrorResponse
// @Router       /api/v1/tenants [get]
func (h *TenantHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := repository.TenantFilter{
		Limit:  20,
		Offset: 0,
	}

	if s := r.URL.Query().Get("landlord_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid landlord_id filter"))
			return
		}
		filter.LandlordID = &id
	}

	if s := r.URL.Query().Get("is_active"); s != "" {
		v := s == "true"
		filter.IsActive = &v
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

	tenants, total, err := h.listUC.Execute(r.Context(), filter)
	if err != nil {
		handleTenantDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTenantListResponse(tenants, total, filter.Limit, filter.Offset))
}

// Update godoc
// @Summary      Update a tenant
// @Description  Updates an existing tenant by ID
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id    path      string                   true  "Tenant ID (UUID)"
// @Param        body  body      dto.UpdateTenantRequest  true  "Updated tenant data"
// @Success      200   {object}  dto.TenantResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      404   {object}  apperror.ErrorResponse
// @Failure      422   {object}  apperror.ErrorResponse
// @Failure      500   {object}  apperror.ErrorResponse
// @Router       /api/v1/tenants/{id} [put]
func (h *TenantHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid tenant ID format"))
		return
	}

	var req dto.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	tenant := req.ToEntity()

	if err := h.updateUC.Execute(r.Context(), id, tenant); err != nil {
		handleTenantDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleTenantDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTenantResponse(updated))
}

// Deactivate godoc
// @Summary      Deactivate a tenant
// @Description  Sets a tenant's is_active to false
// @Tags         tenants
// @Produce      json
// @Param        id   path      string  true  "Tenant ID (UUID)"
// @Success      200  {object}  dto.TenantResponse
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/tenants/{id}/deactivate [put]
func (h *TenantHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid tenant ID format"))
		return
	}

	if err := h.deactivateUC.Execute(r.Context(), id); err != nil {
		handleTenantDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleTenantDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTenantResponse(updated))
}

// Delete godoc
// @Summary      Delete a tenant
// @Description  Soft-deletes a tenant by ID
// @Tags         tenants
// @Produce      json
// @Param        id   path      string  true  "Tenant ID (UUID)"
// @Success      204  "No Content"
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/tenants/{id} [delete]
func (h *TenantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid tenant ID format"))
		return
	}

	if err := h.deleteUC.Execute(r.Context(), id); err != nil {
		handleTenantDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleTenantDomainError(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		apperror.WriteError(w, appErr)
		return
	}

	if errors.Is(err, entity.ErrTenantFullNameRequired) ||
		errors.Is(err, entity.ErrTenantFullNameTooLong) ||
		errors.Is(err, entity.ErrTenantContactRequired) {
		apperror.WriteError(w, apperror.NewValidation(err.Error()))
		return
	}

	if errors.Is(err, entity.ErrTenantEmailExists) {
		apperror.WriteError(w, apperror.NewConflict(err.Error()))
		return
	}

	slog.Error("unhandled error in tenant handler", "error", err)
	apperror.WriteError(w, apperror.NewInternal(err))
}

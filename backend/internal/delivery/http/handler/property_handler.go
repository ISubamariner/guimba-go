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
	propertyuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/property"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/validator"
)

// PropertyHandler handles HTTP requests for properties.
type PropertyHandler struct {
	createUC     *propertyuc.CreatePropertyUseCase
	getUC        *propertyuc.GetPropertyUseCase
	listUC       *propertyuc.ListPropertiesUseCase
	updateUC     *propertyuc.UpdatePropertyUseCase
	deactivateUC *propertyuc.DeactivatePropertyUseCase
	deleteUC     *propertyuc.DeletePropertyUseCase
}

// NewPropertyHandler creates a new PropertyHandler.
func NewPropertyHandler(
	createUC *propertyuc.CreatePropertyUseCase,
	getUC *propertyuc.GetPropertyUseCase,
	listUC *propertyuc.ListPropertiesUseCase,
	updateUC *propertyuc.UpdatePropertyUseCase,
	deactivateUC *propertyuc.DeactivatePropertyUseCase,
	deleteUC *propertyuc.DeletePropertyUseCase,
) *PropertyHandler {
	return &PropertyHandler{
		createUC:     createUC,
		getUC:        getUC,
		listUC:       listUC,
		updateUC:     updateUC,
		deactivateUC: deactivateUC,
		deleteUC:     deleteUC,
	}
}

// Create godoc
// @Summary      Create a property
// @Description  Creates a new property belonging to the authenticated owner
// @Tags         properties
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreatePropertyRequest  true  "Property data"
// @Success      201   {object}  dto.PropertyResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      409   {object}  apperror.ErrorResponse
// @Failure      422   {object}  apperror.ErrorResponse
// @Failure      500   {object}  apperror.ErrorResponse
// @Router       /api/v1/properties [post]
func (h *PropertyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	ownerID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))

	property, err := req.ToEntity(ownerID)
	if err != nil {
		handlePropertyDomainError(w, err)
		return
	}

	if err := h.createUC.Execute(r.Context(), property); err != nil {
		handlePropertyDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewPropertyResponse(property))
}

// Get godoc
// @Summary      Get a property
// @Description  Retrieves a property by ID
// @Tags         properties
// @Produce      json
// @Param        id   path      string  true  "Property ID (UUID)"
// @Success      200  {object}  dto.PropertyResponse
// @Failure      400  {object}  apperror.ErrorResponse
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/properties/{id} [get]
func (h *PropertyHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid property ID format"))
		return
	}

	property, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handlePropertyDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewPropertyResponse(property))
}

// List godoc
// @Summary      List properties
// @Description  Returns a paginated list of properties with optional filtering
// @Tags         properties
// @Produce      json
// @Param        owner_id             query     string  false  "Filter by owner ID"
// @Param        is_active            query     bool    false  "Filter by active status"
// @Param        is_available_for_rent query    bool    false  "Filter by rent availability"
// @Param        property_type        query     string  false  "Filter by property type"
// @Param        search               query     string  false  "Search by name, property_code"
// @Param        limit                query     int     false  "Page size (default 20, max 100)"
// @Param        offset               query     int     false  "Offset (default 0)"
// @Success      200                  {object}  dto.PropertyListResponse
// @Failure      500                  {object}  apperror.ErrorResponse
// @Router       /api/v1/properties [get]
func (h *PropertyHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := repository.PropertyFilter{
		Limit:  20,
		Offset: 0,
	}

	if s := r.URL.Query().Get("owner_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid owner_id filter"))
			return
		}
		filter.OwnerID = &id
	}

	if s := r.URL.Query().Get("is_active"); s != "" {
		v := s == "true"
		filter.IsActive = &v
	}

	if s := r.URL.Query().Get("is_available_for_rent"); s != "" {
		v := s == "true"
		filter.IsAvailableForRent = &v
	}

	if s := r.URL.Query().Get("property_type"); s != "" {
		filter.PropertyType = &s
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

	properties, total, err := h.listUC.Execute(r.Context(), filter)
	if err != nil {
		handlePropertyDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewPropertyListResponse(properties, total, filter.Limit, filter.Offset))
}

// Update godoc
// @Summary      Update a property
// @Description  Updates an existing property by ID
// @Tags         properties
// @Accept       json
// @Produce      json
// @Param        id    path      string                     true  "Property ID (UUID)"
// @Param        body  body      dto.UpdatePropertyRequest  true  "Updated property data"
// @Success      200   {object}  dto.PropertyResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      404   {object}  apperror.ErrorResponse
// @Failure      422   {object}  apperror.ErrorResponse
// @Failure      500   {object}  apperror.ErrorResponse
// @Router       /api/v1/properties/{id} [put]
func (h *PropertyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid property ID format"))
		return
	}

	var req dto.UpdatePropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	property := req.ToEntity()

	if err := h.updateUC.Execute(r.Context(), id, property); err != nil {
		handlePropertyDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handlePropertyDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewPropertyResponse(updated))
}

// Deactivate godoc
// @Summary      Deactivate a property
// @Description  Sets a property's is_active to false
// @Tags         properties
// @Produce      json
// @Param        id   path      string  true  "Property ID (UUID)"
// @Success      200  {object}  dto.PropertyResponse
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/properties/{id}/deactivate [put]
func (h *PropertyHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid property ID format"))
		return
	}

	if err := h.deactivateUC.Execute(r.Context(), id); err != nil {
		handlePropertyDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handlePropertyDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewPropertyResponse(updated))
}

// Delete godoc
// @Summary      Delete a property
// @Description  Soft-deletes a property by ID
// @Tags         properties
// @Produce      json
// @Param        id   path      string  true  "Property ID (UUID)"
// @Success      204  "No Content"
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/properties/{id} [delete]
func (h *PropertyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid property ID format"))
		return
	}

	if err := h.deleteUC.Execute(r.Context(), id); err != nil {
		handlePropertyDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handlePropertyDomainError(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		apperror.WriteError(w, appErr)
		return
	}

	if errors.Is(err, entity.ErrPropertyNameRequired) ||
		errors.Is(err, entity.ErrPropertyNameTooLong) ||
		errors.Is(err, entity.ErrPropertyCodeRequired) ||
		errors.Is(err, entity.ErrPropertySizeRequired) {
		apperror.WriteError(w, apperror.NewValidation(err.Error()))
		return
	}

	if errors.Is(err, entity.ErrPropertyCodeExists) {
		apperror.WriteError(w, apperror.NewConflict(err.Error()))
		return
	}

	slog.Error("unhandled error in property handler", "error", err)
	apperror.WriteError(w, apperror.NewInternal(err))
}

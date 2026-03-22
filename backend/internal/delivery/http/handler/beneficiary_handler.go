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
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	beneficiaryuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/beneficiary"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/validator"
)

// BeneficiaryHandler handles HTTP requests for beneficiaries.
type BeneficiaryHandler struct {
	createUC            *beneficiaryuc.CreateBeneficiaryUseCase
	getUC               *beneficiaryuc.GetBeneficiaryUseCase
	listUC              *beneficiaryuc.ListBeneficiariesUseCase
	updateUC            *beneficiaryuc.UpdateBeneficiaryUseCase
	deleteUC            *beneficiaryuc.DeleteBeneficiaryUseCase
	enrollInProgramUC   *beneficiaryuc.EnrollInProgramUseCase
	removeFromProgramUC *beneficiaryuc.RemoveFromProgramUseCase
}

// NewBeneficiaryHandler creates a new BeneficiaryHandler.
func NewBeneficiaryHandler(
	createUC *beneficiaryuc.CreateBeneficiaryUseCase,
	getUC *beneficiaryuc.GetBeneficiaryUseCase,
	listUC *beneficiaryuc.ListBeneficiariesUseCase,
	updateUC *beneficiaryuc.UpdateBeneficiaryUseCase,
	deleteUC *beneficiaryuc.DeleteBeneficiaryUseCase,
	enrollInProgramUC *beneficiaryuc.EnrollInProgramUseCase,
	removeFromProgramUC *beneficiaryuc.RemoveFromProgramUseCase,
) *BeneficiaryHandler {
	return &BeneficiaryHandler{
		createUC:            createUC,
		getUC:               getUC,
		listUC:              listUC,
		updateUC:            updateUC,
		deleteUC:            deleteUC,
		enrollInProgramUC:   enrollInProgramUC,
		removeFromProgramUC: removeFromProgramUC,
	}
}

// Create godoc
// @Summary      Create a beneficiary
// @Description  Creates a new beneficiary
// @Tags         beneficiaries
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateBeneficiaryRequest  true  "Beneficiary data"
// @Success      201   {object}  dto.BeneficiaryResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      422   {object}  apperror.ErrorResponse
// @Failure      500   {object}  apperror.ErrorResponse
// @Router       /api/v1/beneficiaries [post]
func (h *BeneficiaryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBeneficiaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	beneficiary, err := req.ToEntity()
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest(err.Error()))
		return
	}

	if err := h.createUC.Execute(r.Context(), beneficiary); err != nil {
		handleBeneficiaryDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewBeneficiaryResponse(beneficiary))
}

// Get godoc
// @Summary      Get a beneficiary
// @Description  Retrieves a beneficiary by ID with program enrollments
// @Tags         beneficiaries
// @Produce      json
// @Param        id   path      string  true  "Beneficiary ID (UUID)"
// @Success      200  {object}  dto.BeneficiaryResponse
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/beneficiaries/{id} [get]
func (h *BeneficiaryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid beneficiary ID format"))
		return
	}

	beneficiary, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleBeneficiaryDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewBeneficiaryResponse(beneficiary))
}

// List godoc
// @Summary      List beneficiaries
// @Description  Returns a paginated list of beneficiaries with optional filtering
// @Tags         beneficiaries
// @Produce      json
// @Param        status      query     string  false  "Filter by status (active, inactive, suspended)"
// @Param        program_id  query     string  false  "Filter by program ID"
// @Param        search      query     string  false  "Search by name"
// @Param        limit       query     int     false  "Page size (default 20, max 100)"
// @Param        offset      query     int     false  "Offset (default 0)"
// @Success      200         {object}  dto.BeneficiaryListResponse
// @Failure      500         {object}  apperror.ErrorResponse
// @Router       /api/v1/beneficiaries [get]
func (h *BeneficiaryHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := repository.BeneficiaryFilter{
		Limit:  20,
		Offset: 0,
	}

	if s := r.URL.Query().Get("status"); s != "" {
		status := entity.BeneficiaryStatus(s)
		if !status.IsValid() {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid status filter"))
			return
		}
		filter.Status = &status
	}

	if s := r.URL.Query().Get("program_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid program_id filter"))
			return
		}
		filter.ProgramID = &id
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

	beneficiaries, total, err := h.listUC.Execute(r.Context(), filter)
	if err != nil {
		handleBeneficiaryDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewBeneficiaryListResponse(beneficiaries, total, filter.Limit, filter.Offset))
}

// Update godoc
// @Summary      Update a beneficiary
// @Description  Updates an existing beneficiary by ID
// @Tags         beneficiaries
// @Accept       json
// @Produce      json
// @Param        id    path      string                        true  "Beneficiary ID (UUID)"
// @Param        body  body      dto.UpdateBeneficiaryRequest  true  "Updated beneficiary data"
// @Success      200   {object}  dto.BeneficiaryResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      404   {object}  apperror.ErrorResponse
// @Failure      422   {object}  apperror.ErrorResponse
// @Failure      500   {object}  apperror.ErrorResponse
// @Router       /api/v1/beneficiaries/{id} [put]
func (h *BeneficiaryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid beneficiary ID format"))
		return
	}

	var req dto.UpdateBeneficiaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	beneficiary, err := req.ToEntity()
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest(err.Error()))
		return
	}

	if err := h.updateUC.Execute(r.Context(), id, beneficiary); err != nil {
		handleBeneficiaryDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleBeneficiaryDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewBeneficiaryResponse(updated))
}

// Delete godoc
// @Summary      Delete a beneficiary
// @Description  Soft-deletes a beneficiary by ID
// @Tags         beneficiaries
// @Produce      json
// @Param        id   path      string  true  "Beneficiary ID (UUID)"
// @Success      204  "No Content"
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/beneficiaries/{id} [delete]
func (h *BeneficiaryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid beneficiary ID format"))
		return
	}

	if err := h.deleteUC.Execute(r.Context(), id); err != nil {
		handleBeneficiaryDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EnrollInProgram godoc
// @Summary      Enroll beneficiary in program
// @Description  Enrolls a beneficiary in a social program
// @Tags         beneficiaries
// @Accept       json
// @Produce      json
// @Param        id    path      string                   true  "Beneficiary ID (UUID)"
// @Param        body  body      dto.EnrollProgramRequest true  "Program enrollment"
// @Success      200   {object}  dto.BeneficiaryResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      404   {object}  apperror.ErrorResponse
// @Failure      409   {object}  apperror.ErrorResponse
// @Failure      500   {object}  apperror.ErrorResponse
// @Router       /api/v1/beneficiaries/{id}/programs [post]
func (h *BeneficiaryHandler) EnrollInProgram(w http.ResponseWriter, r *http.Request) {
	beneficiaryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid beneficiary ID format"))
		return
	}

	var req dto.EnrollProgramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	programID, err := uuid.Parse(req.ProgramID)
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid program ID format"))
		return
	}

	if err := h.enrollInProgramUC.Execute(r.Context(), beneficiaryID, programID); err != nil {
		handleBeneficiaryDomainError(w, err)
		return
	}

	// Return updated beneficiary with enrollments
	updated, err := h.getUC.Execute(r.Context(), beneficiaryID)
	if err != nil {
		handleBeneficiaryDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewBeneficiaryResponse(updated))
}

// RemoveFromProgram godoc
// @Summary      Remove beneficiary from program
// @Description  Removes a beneficiary from a social program
// @Tags         beneficiaries
// @Produce      json
// @Param        id         path      string  true  "Beneficiary ID (UUID)"
// @Param        programId  path      string  true  "Program ID (UUID)"
// @Success      204  "No Content"
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/beneficiaries/{id}/programs/{programId} [delete]
func (h *BeneficiaryHandler) RemoveFromProgram(w http.ResponseWriter, r *http.Request) {
	beneficiaryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid beneficiary ID format"))
		return
	}

	programID, err := uuid.Parse(chi.URLParam(r, "programId"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid program ID format"))
		return
	}

	if err := h.removeFromProgramUC.Execute(r.Context(), beneficiaryID, programID); err != nil {
		handleBeneficiaryDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleBeneficiaryDomainError maps domain/app errors to HTTP error responses.
func handleBeneficiaryDomainError(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		apperror.WriteError(w, appErr)
		return
	}

	// Check for domain entity validation errors
	if errors.Is(err, entity.ErrBeneficiaryFullNameRequired) ||
		errors.Is(err, entity.ErrBeneficiaryFullNameTooLong) ||
		errors.Is(err, entity.ErrBeneficiaryInvalidStatus) ||
		errors.Is(err, entity.ErrBeneficiaryContactRequired) {
		apperror.WriteError(w, apperror.NewValidation(err.Error()))
		return
	}

	if errors.Is(err, entity.ErrBeneficiaryAlreadyEnrolled) {
		apperror.WriteError(w, apperror.NewConflict(err.Error()))
		return
	}

	if errors.Is(err, entity.ErrBeneficiaryNotEnrolled) {
		apperror.WriteError(w, apperror.NewNotFoundMsg(err.Error()))
		return
	}

	slog.Error("unhandled error in beneficiary handler", "error", err)
	apperror.WriteError(w, apperror.NewInternal(err))
}

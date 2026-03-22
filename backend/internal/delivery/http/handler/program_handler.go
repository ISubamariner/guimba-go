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
	programuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/program"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/validator"
)

// ProgramHandler handles HTTP requests for programs.
type ProgramHandler struct {
	createUC *programuc.CreateProgramUseCase
	getUC    *programuc.GetProgramUseCase
	listUC   *programuc.ListProgramsUseCase
	updateUC *programuc.UpdateProgramUseCase
	deleteUC *programuc.DeleteProgramUseCase
}

// NewProgramHandler creates a new ProgramHandler.
func NewProgramHandler(
	createUC *programuc.CreateProgramUseCase,
	getUC *programuc.GetProgramUseCase,
	listUC *programuc.ListProgramsUseCase,
	updateUC *programuc.UpdateProgramUseCase,
	deleteUC *programuc.DeleteProgramUseCase,
) *ProgramHandler {
	return &ProgramHandler{
		createUC: createUC,
		getUC:    getUC,
		listUC:   listUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
	}
}

// Create godoc
// @Summary      Create a program
// @Description  Creates a new social protection program
// @Tags         programs
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateProgramRequest  true  "Program data"
// @Success      201   {object}  dto.ProgramResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      422   {object}  apperror.ErrorResponse
// @Failure      500   {object}  apperror.ErrorResponse
// @Router       /api/v1/programs [post]
func (h *ProgramHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProgramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	program, err := req.ToEntity()
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest(err.Error()))
		return
	}

	if err := h.createUC.Execute(r.Context(), program); err != nil {
		handleDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewProgramResponse(program))
}

// Get godoc
// @Summary      Get a program
// @Description  Retrieves a program by its ID
// @Tags         programs
// @Produce      json
// @Param        id   path      string  true  "Program ID (UUID)"
// @Success      200  {object}  dto.ProgramResponse
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/programs/{id} [get]
func (h *ProgramHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid program ID format"))
		return
	}

	program, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewProgramResponse(program))
}

// List godoc
// @Summary      List programs
// @Description  Returns a paginated list of programs with optional filtering
// @Tags         programs
// @Produce      json
// @Param        status  query     string  false  "Filter by status (active, inactive, closed)"
// @Param        search  query     string  false  "Search by name"
// @Param        limit   query     int     false  "Page size (default 20, max 100)"
// @Param        offset  query     int     false  "Offset (default 0)"
// @Success      200     {object}  dto.ProgramListResponse
// @Failure      500     {object}  apperror.ErrorResponse
// @Router       /api/v1/programs [get]
func (h *ProgramHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := repository.ProgramFilter{
		Limit:  20,
		Offset: 0,
	}

	if s := r.URL.Query().Get("status"); s != "" {
		status := entity.ProgramStatus(s)
		if !status.IsValid() {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid status filter"))
			return
		}
		filter.Status = &status
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

	programs, total, err := h.listUC.Execute(r.Context(), filter)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewProgramListResponse(programs, total, filter.Limit, filter.Offset))
}

// Update godoc
// @Summary      Update a program
// @Description  Updates an existing program by ID
// @Tags         programs
// @Accept       json
// @Produce      json
// @Param        id    path      string                    true  "Program ID (UUID)"
// @Param        body  body      dto.UpdateProgramRequest  true  "Updated program data"
// @Success      200   {object}  dto.ProgramResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      404   {object}  apperror.ErrorResponse
// @Failure      422   {object}  apperror.ErrorResponse
// @Failure      500   {object}  apperror.ErrorResponse
// @Router       /api/v1/programs/{id} [put]
func (h *ProgramHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid program ID format"))
		return
	}

	var req dto.UpdateProgramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	program, err := req.ToEntity()
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest(err.Error()))
		return
	}

	if err := h.updateUC.Execute(r.Context(), id, program); err != nil {
		handleDomainError(w, err)
		return
	}

	// Re-fetch the updated program for the response
	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewProgramResponse(updated))
}

// Delete godoc
// @Summary      Delete a program
// @Description  Soft-deletes a program by ID
// @Tags         programs
// @Produce      json
// @Param        id   path      string  true  "Program ID (UUID)"
// @Success      204  "No Content"
// @Failure      404  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/programs/{id} [delete]
func (h *ProgramHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid program ID format"))
		return
	}

	if err := h.deleteUC.Execute(r.Context(), id); err != nil {
		handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleDomainError maps domain/app errors to HTTP error responses.
func handleDomainError(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		apperror.WriteError(w, appErr)
		return
	}

	// Check for domain entity validation errors
	if errors.Is(err, entity.ErrProgramNameRequired) ||
		errors.Is(err, entity.ErrProgramNameTooLong) ||
		errors.Is(err, entity.ErrProgramInvalidStatus) ||
		errors.Is(err, entity.ErrProgramEndBeforeStart) {
		apperror.WriteError(w, apperror.NewValidation(err.Error()))
		return
	}

	slog.Error("unhandled error in handler", "error", err)
	apperror.WriteError(w, apperror.NewInternal(err))
}

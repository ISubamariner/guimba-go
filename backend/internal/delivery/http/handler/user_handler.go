package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	useruc "github.com/ISubamariner/guimba-go/backend/internal/usecase/user"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/validator"
)

// UserHandler handles user management requests.
type UserHandler struct {
	listUC       *useruc.ListUsersUseCase
	updateUC     *useruc.UpdateUserUseCase
	deleteUC     *useruc.DeleteUserUseCase
	assignRoleUC *useruc.AssignRoleUseCase
	removeRoleUC *useruc.RemoveRoleUseCase
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(
	listUC *useruc.ListUsersUseCase,
	updateUC *useruc.UpdateUserUseCase,
	deleteUC *useruc.DeleteUserUseCase,
	assignRoleUC *useruc.AssignRoleUseCase,
	removeRoleUC *useruc.RemoveRoleUseCase,
) *UserHandler {
	return &UserHandler{
		listUC:       listUC,
		updateUC:     updateUC,
		deleteUC:     deleteUC,
		assignRoleUC: assignRoleUC,
		removeRoleUC: removeRoleUC,
	}
}

// List godoc
// @Summary      List users
// @Description  Returns a paginated list of users (admin only)
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        search    query     string  false  "Search by name or email"
// @Param        is_active query     bool    false  "Filter by active status"
// @Param        role      query     string  false  "Filter by role name"
// @Param        limit     query     int     false  "Page size (default 20, max 100)"
// @Param        offset    query     int     false  "Offset (default 0)"
// @Success      200       {object}  dto.UserListResponse
// @Failure      403       {object}  apperror.ErrorResponse
// @Router       /api/v1/users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := repository.UserFilter{Limit: 20, Offset: 0}

	if s := r.URL.Query().Get("search"); s != "" {
		filter.Search = &s
	}
	if s := r.URL.Query().Get("is_active"); s != "" {
		if v, err := strconv.ParseBool(s); err == nil {
			filter.IsActive = &v
		}
	}
	if s := r.URL.Query().Get("role"); s != "" {
		filter.RoleName = &s
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

	users, total, err := h.listUC.Execute(r.Context(), filter)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewUserListResponse(users, total, filter.Limit, filter.Offset))
}

// Update godoc
// @Summary      Update a user
// @Description  Updates user profile fields (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                true  "User ID (UUID)"
// @Param        body  body      dto.UpdateUserRequest true  "Updated user data"
// @Success      200   {object}  dto.UserResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      404   {object}  apperror.ErrorResponse
// @Router       /api/v1/users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid user ID format"))
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	user, err := h.updateUC.Execute(r.Context(), id, req.FullName, req.IsActive)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewUserResponse(user))
}

// Delete godoc
// @Summary      Delete a user
// @Description  Soft-deletes a user (admin only)
// @Tags         users
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      204  "No Content"
// @Failure      404  {object}  apperror.ErrorResponse
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid user ID format"))
		return
	}

	if err := h.deleteUC.Execute(r.Context(), id); err != nil {
		handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignRole godoc
// @Summary      Assign a role to a user
// @Description  Assigns a role to a user (admin only)
// @Tags         users
// @Accept       json
// @Security     BearerAuth
// @Param        id    path      string              true  "User ID (UUID)"
// @Param        body  body      dto.AssignRoleRequest true "Role to assign"
// @Success      204   "No Content"
// @Failure      404   {object}  apperror.ErrorResponse
// @Router       /api/v1/users/{id}/roles [post]
func (h *UserHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid user ID format"))
		return
	}

	var req dto.AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	if err := h.assignRoleUC.Execute(r.Context(), userID, req.RoleID); err != nil {
		handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveRole godoc
// @Summary      Remove a role from a user
// @Description  Removes a role from a user (admin only)
// @Tags         users
// @Security     BearerAuth
// @Param        id       path      string  true  "User ID (UUID)"
// @Param        roleId   path      string  true  "Role ID (UUID)"
// @Success      204      "No Content"
// @Failure      404      {object}  apperror.ErrorResponse
// @Router       /api/v1/users/{id}/roles/{roleId} [delete]
func (h *UserHandler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid user ID format"))
		return
	}

	roleID, err := uuid.Parse(chi.URLParam(r, "roleId"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid role ID format"))
		return
	}

	if err := h.removeRoleUC.Execute(r.Context(), userID, roleID); err != nil {
		handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

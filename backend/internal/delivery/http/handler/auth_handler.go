package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/cache"
	authuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/auth"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
	"github.com/ISubamariner/guimba-go/backend/pkg/validator"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/google/uuid"
)

// AuthHandler handles authentication requests.
type AuthHandler struct {
	registerUC       *authuc.RegisterUseCase
	loginUC          *authuc.LoginUseCase
	refreshUC        *authuc.RefreshTokenUseCase
	profileUC        *authuc.GetProfileUseCase
	changePasswordUC *authuc.ChangePasswordUseCase
	jwt              *auth.JWTManager
	blocklist        *cache.TokenBlocklist
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	registerUC *authuc.RegisterUseCase,
	loginUC *authuc.LoginUseCase,
	refreshUC *authuc.RefreshTokenUseCase,
	profileUC *authuc.GetProfileUseCase,
	changePasswordUC *authuc.ChangePasswordUseCase,
	jwt *auth.JWTManager,
	blocklist *cache.TokenBlocklist,
) *AuthHandler {
	return &AuthHandler{
		registerUC:       registerUC,
		loginUC:          loginUC,
		refreshUC:        refreshUC,
		profileUC:        profileUC,
		changePasswordUC: changePasswordUC,
		jwt:              jwt,
		blocklist:        blocklist,
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account and returns JWT tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterRequest  true  "Registration data"
// @Success      201   {object}  dto.AuthResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      409   {object}  apperror.ErrorResponse
// @Failure      422   {object}  apperror.ErrorResponse
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	user, tokens, err := h.registerUC.Execute(r.Context(), req.Email, req.FullName, req.Password)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.AuthResponse{
		User:         dto.NewUserResponse(user),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// Login godoc
// @Summary      Login
// @Description  Authenticates a user and returns JWT tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Login credentials"
// @Success      200   {object}  dto.AuthResponse
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      401   {object}  apperror.ErrorResponse
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	user, tokens, err := h.loginUC.Execute(r.Context(), req.Email, req.Password)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.AuthResponse{
		User:         dto.NewUserResponse(user),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// Refresh godoc
// @Summary      Refresh tokens
// @Description  Exchanges a refresh token for a new token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RefreshRequest  true  "Refresh token"
// @Success      200   {object}  dto.TokenResponse
// @Failure      401   {object}  apperror.ErrorResponse
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	tokens, err := h.refreshUC.Execute(r.Context(), req.RefreshToken)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// Me godoc
// @Summary      Get current user profile
// @Description  Returns the authenticated user's profile
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.UserResponse
// @Failure      401  {object}  apperror.ErrorResponse
// @Router       /api/v1/auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User not authenticated"))
		return
	}

	user, err := h.profileUC.Execute(r.Context(), userID)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewUserResponse(user))
}

// Logout godoc
// @Summary      Logout
// @Description  Revokes the current access token
// @Tags         auth
// @Security     BearerAuth
// @Success      204  "No Content"
// @Failure      401  {object}  apperror.ErrorResponse
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("Authorization")
	if header == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	parts := splitBearer(header)
	if parts == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	claims, err := h.jwt.ValidateToken(parts)
	if err == nil && claims.ID != "" {
		remaining := time.Until(claims.ExpiresAt.Time)
		if remaining > 0 {
			_ = h.blocklist.Block(r.Context(), claims.ID, remaining)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// ChangePassword godoc
// @Summary      Change password
// @Description  Changes the authenticated user's password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  dto.ChangePasswordRequest  true  "Password change data"
// @Success      204   "No Content"
// @Failure      400   {object}  apperror.ErrorResponse
// @Failure      401   {object}  apperror.ErrorResponse
// @Failure      422   {object}  apperror.ErrorResponse
// @Router       /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User not authenticated"))
		return
	}

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	if err := h.changePasswordUC.Execute(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func splitBearer(header string) string {
	const prefix = "Bearer "
	if len(header) > len(prefix) && header[:len(prefix)] == prefix {
		return header[len(prefix):]
	}
	return ""
}

package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	audituc "github.com/ISubamariner/guimba-go/backend/internal/usecase/audit"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type AuditHandler struct {
	listUC         *audituc.ListAuditLogsUseCase
	listLandlordUC *audituc.ListLandlordAuditLogsUseCase
}

func NewAuditHandler(listUC *audituc.ListAuditLogsUseCase, listLandlordUC *audituc.ListLandlordAuditLogsUseCase) *AuditHandler {
	return &AuditHandler{listUC: listUC, listLandlordUC: listLandlordUC}
}

// List godoc
// @Summary      List audit logs
// @Description  Returns audit log entries with optional filters (admin/auditor only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        user_id       query  string  false  "Filter by user ID"
// @Param        action        query  string  false  "Filter by action"
// @Param        resource_type query  string  false  "Filter by resource type"
// @Param        success       query  bool    false  "Filter by success"
// @Param        from_date     query  string  false  "Filter from date (RFC3339)"
// @Param        to_date       query  string  false  "Filter to date (RFC3339)"
// @Param        limit         query  int     false  "Limit (default 20, max 100)"
// @Param        offset        query  int     false  "Offset"
// @Success      200  {object}  dto.AuditListResponse
// @Failure      400  {object}  apperror.ErrorResponse
// @Router       /audit [get]
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseAuditFilter(r)
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest(err.Error()))
		return
	}

	entries, total, err := h.listUC.Execute(r.Context(), filter)
	if err != nil {
		apperror.WriteError(w, apperror.NewInternal(err))
		return
	}

	resp := dto.NewAuditListResponse(entries, total, filter.Limit, filter.Offset)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// LandlordList godoc
// @Summary      List landlord-scoped audit logs
// @Description  Returns audit log entries scoped to the authenticated landlord
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        action        query  string  false  "Filter by action"
// @Param        resource_type query  string  false  "Filter by resource type"
// @Param        from_date     query  string  false  "Filter from date (RFC3339)"
// @Param        to_date       query  string  false  "Filter to date (RFC3339)"
// @Param        limit         query  int     false  "Limit (default 20, max 100)"
// @Param        offset        query  int     false  "Offset"
// @Success      200  {object}  dto.AuditListResponse
// @Failure      400  {object}  apperror.ErrorResponse
// @Router       /audit/landlord [get]
func (h *AuditHandler) LandlordList(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(middleware.AuthUserIDKey).(string)
	if !ok || userIDStr == "" {
		apperror.WriteError(w, apperror.NewUnauthorized("Missing user ID"))
		return
	}
	landlordID, err := uuid.Parse(userIDStr)
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid user ID"))
		return
	}

	filter, err := parseAuditFilter(r)
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest(err.Error()))
		return
	}

	entries, total, err := h.listLandlordUC.Execute(r.Context(), landlordID, filter)
	if err != nil {
		apperror.WriteError(w, apperror.NewInternal(err))
		return
	}

	resp := dto.NewAuditListResponse(entries, total, filter.Limit, filter.Offset)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func parseAuditFilter(r *http.Request) (repository.AuditFilter, error) {
	filter := repository.AuditFilter{}

	if s := r.URL.Query().Get("user_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			return filter, err
		}
		filter.UserID = &id
	}
	if s := r.URL.Query().Get("action"); s != "" {
		filter.Action = &s
	}
	if s := r.URL.Query().Get("resource_type"); s != "" {
		filter.ResourceType = &s
	}
	if s := r.URL.Query().Get("success"); s != "" {
		v := s == "true"
		filter.Success = &v
	}
	if s := r.URL.Query().Get("from_date"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return filter, err
		}
		filter.FromDate = &t
	}
	if s := r.URL.Query().Get("to_date"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return filter, err
		}
		filter.ToDate = &t
	}
	if s := r.URL.Query().Get("limit"); s != "" {
		v, _ := strconv.Atoi(s)
		filter.Limit = v
	}
	if s := r.URL.Query().Get("offset"); s != "" {
		v, _ := strconv.Atoi(s)
		filter.Offset = v
	}

	return filter, nil
}

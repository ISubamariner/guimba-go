package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	dashuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/dashboard"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// DashboardHandler handles HTTP requests for dashboard endpoints.
type DashboardHandler struct {
	getStatsUC             *dashuc.GetStatsUseCase
	getRecentActivitiesUC  *dashuc.GetRecentActivitiesUseCase
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(
	getStatsUC *dashuc.GetStatsUseCase,
	getRecentActivitiesUC *dashuc.GetRecentActivitiesUseCase,
) *DashboardHandler {
	return &DashboardHandler{
		getStatsUC:            getStatsUC,
		getRecentActivitiesUC: getRecentActivitiesUC,
	}
}

// Stats godoc
// @Summary      Get dashboard statistics
// @Description  Returns aggregated statistics for the authenticated landlord's portfolio
// @Tags         dashboard
// @Produce      json
// @Success      200  {object}  dto.DashboardStatsResponse
// @Failure      401  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/dashboard/stats [get]
func (h *DashboardHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User ID not found in context"))
		return
	}

	stats, err := h.getStatsUC.Execute(r.Context(), userID)
	if err != nil {
		slog.Error("failed to get dashboard stats", "error", err, "user_id", userID)
		apperror.WriteError(w, apperror.NewInternal(err))
		return
	}

	response := dto.DashboardStatsResponse{
		TotalTenants:    stats.TotalTenants,
		TotalProperties: stats.TotalProperties,
		ActiveDebts:     stats.ActiveDebts,
		OverdueDebts:    stats.OverdueDebts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RecentActivities godoc
// @Summary      Get recent activities
// @Description  Returns recent audit activities for the authenticated landlord
// @Tags         dashboard
// @Produce      json
// @Param        limit  query     int  false  "Max number of activities (default 10, max 50)"
// @Success      200    {object}  dto.RecentActivitiesResponse
// @Failure      401    {object}  apperror.ErrorResponse
// @Failure      500    {object}  apperror.ErrorResponse
// @Router       /api/v1/dashboard/recent-activities [get]
func (h *DashboardHandler) RecentActivities(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User ID not found in context"))
		return
	}

	limit := 10
	if s := r.URL.Query().Get("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			limit = v
		}
	}

	activities, err := h.getRecentActivitiesUC.Execute(r.Context(), userID, limit)
	if err != nil {
		slog.Error("failed to get recent activities", "error", err, "user_id", userID)
		apperror.WriteError(w, apperror.NewInternal(err))
		return
	}

	activityResponses := make([]dto.RecentActivityResponse, 0, len(activities))
	for _, activity := range activities {
		activityResponses = append(activityResponses, dto.RecentActivityResponse{
			Action:      activity.Action,
			Description: activity.Description,
			Timestamp:   activity.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	response := dto.RecentActivitiesResponse{
		Data: activityResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

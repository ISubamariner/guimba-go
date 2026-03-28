package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	exportuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/export"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// ExportHandler handles HTTP requests for CSV data export.
type ExportHandler struct {
	exportTenantsUC    *exportuc.ExportTenantsUseCase
	exportPropertiesUC *exportuc.ExportPropertiesUseCase
	exportDebtsUC      *exportuc.ExportDebtsUseCase
}

// NewExportHandler creates a new ExportHandler.
func NewExportHandler(
	exportTenantsUC *exportuc.ExportTenantsUseCase,
	exportPropertiesUC *exportuc.ExportPropertiesUseCase,
	exportDebtsUC *exportuc.ExportDebtsUseCase,
) *ExportHandler {
	return &ExportHandler{
		exportTenantsUC:    exportTenantsUC,
		exportPropertiesUC: exportPropertiesUC,
		exportDebtsUC:      exportDebtsUC,
	}
}

// setCSVHeaders sets the appropriate headers for CSV download.
func setCSVHeaders(w http.ResponseWriter, resource string) {
	filename := fmt.Sprintf("%s_%s.csv", resource, time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
}

// ExportTenants godoc
// @Summary      Export tenants as CSV
// @Description  Downloads all tenants for the authenticated landlord as a CSV file
// @Tags         export
// @Produce      text/csv
// @Success      200  {file}  file  "CSV file"
// @Failure      401  {object}  apperror.ErrorResponse
// @Failure      403  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/export/tenants [get]
func (h *ExportHandler) ExportTenants(w http.ResponseWriter, r *http.Request) {
	landlordID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User not authenticated"))
		return
	}

	setCSVHeaders(w, "tenants")

	if err := h.exportTenantsUC.Execute(r.Context(), landlordID, w); err != nil {
		apperror.WriteError(w, apperror.NewInternal(err))
		return
	}
}

// ExportProperties godoc
// @Summary      Export properties as CSV
// @Description  Downloads all properties for the authenticated landlord as a CSV file
// @Tags         export
// @Produce      text/csv
// @Success      200  {file}  file  "CSV file"
// @Failure      401  {object}  apperror.ErrorResponse
// @Failure      403  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/export/properties [get]
func (h *ExportHandler) ExportProperties(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User not authenticated"))
		return
	}

	setCSVHeaders(w, "properties")

	if err := h.exportPropertiesUC.Execute(r.Context(), ownerID, w); err != nil {
		apperror.WriteError(w, apperror.NewInternal(err))
		return
	}
}

// ExportDebts godoc
// @Summary      Export debts as CSV
// @Description  Downloads all debts for the authenticated landlord as a CSV file
// @Tags         export
// @Produce      text/csv
// @Success      200  {file}  file  "CSV file"
// @Failure      401  {object}  apperror.ErrorResponse
// @Failure      403  {object}  apperror.ErrorResponse
// @Failure      500  {object}  apperror.ErrorResponse
// @Router       /api/v1/export/debts [get]
func (h *ExportHandler) ExportDebts(w http.ResponseWriter, r *http.Request) {
	landlordID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User not authenticated"))
		return
	}

	setCSVHeaders(w, "debts")

	if err := h.exportDebtsUC.Execute(r.Context(), landlordID, w); err != nil {
		apperror.WriteError(w, apperror.NewInternal(err))
		return
	}
}

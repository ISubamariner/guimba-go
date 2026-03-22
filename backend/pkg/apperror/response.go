package apperror

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the JSON shape for all API error responses.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains the error details.
type ErrorBody struct {
	Code    Code     `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

// WriteError writes a structured error response to the HTTP response writer.
func WriteError(w http.ResponseWriter, appErr *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)

	resp := ErrorResponse{
		Error: ErrorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
	}

	json.NewEncoder(w).Encode(resp)
}

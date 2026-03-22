package dto

import "time"

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// NewHealthResponse creates a health response with service statuses.
func NewHealthResponse(services map[string]string) HealthResponse {
	status := "ok"
	for _, v := range services {
		if v != "up" {
			status = "degraded"
			break
		}
	}
	return HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  services,
	}
}

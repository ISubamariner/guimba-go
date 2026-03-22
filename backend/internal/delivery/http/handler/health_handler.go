package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	pgPool      *pgxpool.Pool
	mongoClient *mongo.Client
	redisClient *redis.Client
}

// NewHealthHandler creates a new health check handler.
func NewHealthHandler(pg *pgxpool.Pool, mongo *mongo.Client, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		pgPool:      pg,
		mongoClient: mongo,
		redisClient: redis,
	}
}

// Health godoc
// @Summary      Health check
// @Description  Returns the health status of the API and its dependencies
// @Tags         system
// @Produce      json
// @Success      200  {object}  dto.HealthResponse
// @Router       /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)

	// Check PostgreSQL
	if err := h.pgPool.Ping(ctx); err != nil {
		services["postgres"] = "down"
	} else {
		services["postgres"] = "up"
	}

	// Check MongoDB
	if err := h.mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		services["mongodb"] = "down"
	} else {
		services["mongodb"] = "up"
	}

	// Check Redis
	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		services["redis"] = "down"
	} else {
		services["redis"] = "up"
	}

	resp := dto.NewHealthResponse(services)

	statusCode := http.StatusOK
	if resp.Status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

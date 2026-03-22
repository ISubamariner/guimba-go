package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// NewMongoClient creates a new MongoDB client and verifies connectivity.
func NewMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
	opts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(25).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(5 * time.Minute)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("connecting to MongoDB: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("pinging MongoDB: %w", err)
	}

	slog.Info("connected to MongoDB")
	return client, nil
}

// CloseMongo gracefully disconnects the MongoDB client.
func CloseMongo(ctx context.Context, client *mongo.Client) {
	if client != nil {
		if err := client.Disconnect(ctx); err != nil {
			slog.Error("failed to disconnect MongoDB", "error", err)
			return
		}
		slog.Info("MongoDB connection closed")
	}
}

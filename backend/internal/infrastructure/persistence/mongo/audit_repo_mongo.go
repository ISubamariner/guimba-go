package mongo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type AuditRepoMongo struct {
	collection *mongo.Collection
}

func NewAuditRepoMongo(client *mongo.Client, dbName string) *AuditRepoMongo {
	return &AuditRepoMongo{
		collection: client.Database(dbName).Collection("audit_logs"),
	}
}

type auditDocument struct {
	ID           string         `bson:"_id"`
	UserID       string         `bson:"user_id"`
	UserEmail    string         `bson:"user_email"`
	UserRole     string         `bson:"user_role"`
	Action       string         `bson:"action"`
	ResourceType string         `bson:"resource_type"`
	ResourceID   string         `bson:"resource_id"`
	IPAddress    string         `bson:"ip_address"`
	UserAgent    string         `bson:"user_agent"`
	Endpoint     string         `bson:"endpoint"`
	Method       string         `bson:"method"`
	StatusCode   int            `bson:"status_code"`
	Success      bool           `bson:"success"`
	ErrorMessage *string        `bson:"error_message,omitempty"`
	Metadata     map[string]any `bson:"metadata,omitempty"`
	Timestamp    time.Time      `bson:"timestamp"`
}

func toDocument(e *repository.AuditEntry) auditDocument {
	return auditDocument{
		ID:           e.ID.String(),
		UserID:       e.UserID.String(),
		UserEmail:    e.UserEmail,
		UserRole:     e.UserRole,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID.String(),
		IPAddress:    e.IPAddress,
		UserAgent:    e.UserAgent,
		Endpoint:     e.Endpoint,
		Method:       e.Method,
		StatusCode:   e.StatusCode,
		Success:      e.Success,
		ErrorMessage: e.ErrorMessage,
		Metadata:     e.Metadata,
		Timestamp:    e.Timestamp,
	}
}

func fromDocument(d auditDocument) *repository.AuditEntry {
	id, _ := uuid.Parse(d.ID)
	userID, _ := uuid.Parse(d.UserID)
	resourceID, _ := uuid.Parse(d.ResourceID)
	return &repository.AuditEntry{
		ID:           id,
		UserID:       userID,
		UserEmail:    d.UserEmail,
		UserRole:     d.UserRole,
		Action:       d.Action,
		ResourceType: d.ResourceType,
		ResourceID:   resourceID,
		IPAddress:    d.IPAddress,
		UserAgent:    d.UserAgent,
		Endpoint:     d.Endpoint,
		Method:       d.Method,
		StatusCode:   d.StatusCode,
		Success:      d.Success,
		ErrorMessage: d.ErrorMessage,
		Metadata:     d.Metadata,
		Timestamp:    d.Timestamp,
	}
}

func (r *AuditRepoMongo) Log(ctx context.Context, entry *repository.AuditEntry) error {
	doc := toDocument(entry)
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *AuditRepoMongo) List(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
	bsonFilter := bson.D{}

	if filter.UserID != nil {
		bsonFilter = append(bsonFilter, bson.E{Key: "user_id", Value: filter.UserID.String()})
	}
	if filter.LandlordID != nil {
		lid := filter.LandlordID.String()
		bsonFilter = append(bsonFilter, bson.E{
			Key: "$or",
			Value: bson.A{
				bson.D{{Key: "user_id", Value: lid}},
				bson.D{{Key: "metadata.landlord_id", Value: lid}},
			},
		})
	}
	if filter.Action != nil {
		bsonFilter = append(bsonFilter, bson.E{Key: "action", Value: *filter.Action})
	}
	if filter.ResourceType != nil {
		bsonFilter = append(bsonFilter, bson.E{Key: "resource_type", Value: *filter.ResourceType})
	}
	if filter.Success != nil {
		bsonFilter = append(bsonFilter, bson.E{Key: "success", Value: *filter.Success})
	}
	if filter.FromDate != nil || filter.ToDate != nil {
		dateFilter := bson.D{}
		if filter.FromDate != nil {
			dateFilter = append(dateFilter, bson.E{Key: "$gte", Value: *filter.FromDate})
		}
		if filter.ToDate != nil {
			dateFilter = append(dateFilter, bson.E{Key: "$lte", Value: *filter.ToDate})
		}
		bsonFilter = append(bsonFilter, bson.E{Key: "timestamp", Value: dateFilter})
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		return nil, 0, err
	}

	// Query with pagination and sort
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64(filter.Offset)).
		SetLimit(int64(filter.Limit))

	cursor, err := r.collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var entries []*repository.AuditEntry
	for cursor.Next(ctx) {
		var doc auditDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, err
		}
		entries = append(entries, fromDocument(doc))
	}

	return entries, int(total), cursor.Err()
}

// EnsureIndexes creates indexes for efficient audit log queries.
func (r *AuditRepoMongo) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "resource_type", Value: 1}}},
		{Keys: bson.D{{Key: "action", Value: 1}}},
		{Keys: bson.D{{Key: "timestamp", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "timestamp", Value: -1}}},
		{Keys: bson.D{{Key: "metadata.landlord_id", Value: 1}}},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

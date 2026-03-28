package unit

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	property "github.com/ISubamariner/guimba-go/backend/internal/usecase/property"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

// --- CreateProperty ---

func TestCreateProperty_Success(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByPropertyCodeFn: func(ctx context.Context, code string) (*entity.Property, error) {
			return nil, nil
		},
		CreateFn: func(ctx context.Context, p *entity.Property) error { return nil },
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true, Roles: []entity.Role{{Name: "landlord"}}}, nil
		},
	}

	uc := property.NewCreatePropertyUseCase(repo, userRepo, &mocks.AuditRepositoryMock{})
	p, _ := entity.NewProperty("Farm", "F-001", nil, nil, "LAND", nil, 500.0, uuid.New(), nil, nil)
	err := uc.Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateProperty_OwnerNotFound(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return nil, nil
		},
	}

	uc := property.NewCreatePropertyUseCase(repo, userRepo, &mocks.AuditRepositoryMock{})
	p, _ := entity.NewProperty("Farm", "F-001", nil, nil, "LAND", nil, 500.0, uuid.New(), nil, nil)
	err := uc.Execute(context.Background(), p)
	if err == nil {
		t.Fatal("expected error when owner not found")
	}
}

func TestCreateProperty_DuplicateCode(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByPropertyCodeFn: func(ctx context.Context, code string) (*entity.Property, error) {
			return &entity.Property{ID: uuid.New(), PropertyCode: code}, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true, Roles: []entity.Role{{Name: "landlord"}}}, nil
		},
	}

	uc := property.NewCreatePropertyUseCase(repo, userRepo, &mocks.AuditRepositoryMock{})
	p, _ := entity.NewProperty("Farm", "F-001", nil, nil, "LAND", nil, 500.0, uuid.New(), nil, nil)
	err := uc.Execute(context.Background(), p)
	if err == nil {
		t.Fatal("expected error for duplicate code")
	}
}

// --- GetProperty ---

func TestGetProperty_Success(t *testing.T) {
	propID := uuid.New()
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, Name: "Farm", PropertyCode: "F-001", SizeInSqm: 500, OwnerID: uuid.New()}, nil
		},
	}
	uc := property.NewGetPropertyUseCase(repo)
	result, err := uc.Execute(context.Background(), propID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != propID {
		t.Error("expected ID to match")
	}
}

func TestGetProperty_NotFound(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return nil, nil
		},
	}
	uc := property.NewGetPropertyUseCase(repo)
	_, err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- ListProperties ---

func TestListProperties_Success(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
			return []*entity.Property{{ID: uuid.New(), Name: "Farm", SizeInSqm: 500}}, 1, nil
		},
	}
	uc := property.NewListPropertiesUseCase(repo)
	props, total, err := uc.Execute(context.Background(), repository.PropertyFilter{Limit: 20})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 || len(props) != 1 {
		t.Errorf("expected 1 property, got %d", len(props))
	}
}

func TestListProperties_DefaultLimit(t *testing.T) {
	var captured repository.PropertyFilter
	repo := &mocks.PropertyRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
			captured = filter
			return nil, 0, nil
		},
	}
	uc := property.NewListPropertiesUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), repository.PropertyFilter{Limit: 0})
	if captured.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", captured.Limit)
	}
}

func TestListProperties_MaxLimit(t *testing.T) {
	var captured repository.PropertyFilter
	repo := &mocks.PropertyRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
			captured = filter
			return nil, 0, nil
		},
	}
	uc := property.NewListPropertiesUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), repository.PropertyFilter{Limit: 500})
	if captured.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", captured.Limit)
	}
}

// --- UpdateProperty ---

func TestUpdateProperty_Success(t *testing.T) {
	propID := uuid.New()
	ownerID := uuid.New()
	existing := &entity.Property{ID: propID, Name: "Farm", PropertyCode: "F-001", SizeInSqm: 500, OwnerID: ownerID, IsActive: true}

	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, p *entity.Property) error { return nil },
	}

	uc := property.NewUpdatePropertyUseCase(repo, &mocks.AuditRepositoryMock{})
	updated := &entity.Property{Name: "Updated Farm", PropertyCode: "F-001", SizeInSqm: 600}
	err := uc.Execute(context.Background(), propID, updated)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.OwnerID != ownerID {
		t.Error("expected OwnerID to be preserved")
	}
}

func TestUpdateProperty_NotFound(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return nil, nil
		},
	}
	uc := property.NewUpdatePropertyUseCase(repo, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), uuid.New(), &entity.Property{Name: "X", PropertyCode: "X", SizeInSqm: 100})
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- DeactivateProperty ---

func TestDeactivateProperty_Success(t *testing.T) {
	propID := uuid.New()
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, Name: "Farm", PropertyCode: "F-001", SizeInSqm: 500, IsActive: true}, nil
		},
		UpdateFn: func(ctx context.Context, p *entity.Property) error {
			if p.IsActive {
				t.Error("expected IsActive to be false")
			}
			return nil
		},
	}
	uc := property.NewDeactivatePropertyUseCase(repo, &mocks.DebtRepositoryMock{}, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), propID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeactivateProperty_NotFound(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return nil, nil
		},
	}
	uc := property.NewDeactivatePropertyUseCase(repo, &mocks.DebtRepositoryMock{}, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- DeleteProperty ---

func TestDeleteProperty_Success(t *testing.T) {
	propID := uuid.New()
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, Name: "Farm", PropertyCode: "F-001", SizeInSqm: 500}, nil
		},
		DeleteFn: func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	uc := property.NewDeletePropertyUseCase(repo, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), propID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteProperty_NotFound(t *testing.T) {
	repo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return nil, nil
		},
	}
	uc := property.NewDeletePropertyUseCase(repo, &mocks.AuditRepositoryMock{})
	err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

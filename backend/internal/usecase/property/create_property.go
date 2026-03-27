package property

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type CreatePropertyUseCase struct {
	repo     repository.PropertyRepository
	userRepo repository.UserRepository
}

func NewCreatePropertyUseCase(repo repository.PropertyRepository, userRepo repository.UserRepository) *CreatePropertyUseCase {
	return &CreatePropertyUseCase{repo: repo, userRepo: userRepo}
}

func (uc *CreatePropertyUseCase) Execute(ctx context.Context, property *entity.Property) error {
	if err := property.Validate(); err != nil {
		return err
	}

	owner, err := uc.userRepo.GetByID(ctx, property.OwnerID)
	if err != nil {
		return err
	}
	if owner == nil {
		return apperror.NewNotFound("User", property.OwnerID)
	}

	existing, err := uc.repo.GetByPropertyCode(ctx, property.PropertyCode)
	if err != nil {
		return err
	}
	if existing != nil {
		return apperror.NewConflict(entity.ErrPropertyCodeExists.Error())
	}

	return uc.repo.Create(ctx, property)
}

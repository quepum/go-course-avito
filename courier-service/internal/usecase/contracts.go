package usecase

import (
	"context"
	"courier-service/internal/model"
)

type HealthRepository interface {
	Ping(ctx context.Context) error
}

type ProfileRepository interface {
	Ping(ctx context.Context) error
	GetOneById(ctx context.Context, id int) (*model.ProfileDB, error)
	GetAll(ctx context.Context) ([]model.ProfileDB, error)
	Create(ctx context.Context, courier *model.ProfileDB) (int, error)
	Update(ctx context.Context, courier *model.ProfileDB) error
	Delete(ctx context.Context, id int) error
}

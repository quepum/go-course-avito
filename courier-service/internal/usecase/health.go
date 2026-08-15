package usecase

import (
	"context"
	"courier-service/internal/model"
)

type HealthUseCase struct {
	repository HealthRepository
}

func NewHealthUseCase(repository HealthRepository) *HealthUseCase {
	return &HealthUseCase{repository: repository}
}

func (u *HealthUseCase) Ping(ctx context.Context) model.PingResult {
	return model.PingResult{
		Status: model.PingStatusUp,
	}
}

func (u *HealthUseCase) HealthCheck(ctx context.Context) error {
	return u.repository.Ping(ctx)
}

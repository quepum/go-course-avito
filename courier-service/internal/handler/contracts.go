package handler

import (
	"context"
	"courier-service/internal/model"
)

type ProfileUseCase interface {
	GetProfile(ctx context.Context, id int) (*model.Profile, error)
	GetAllProfiles(ctx context.Context) ([]model.Profile, error)
	CreateProfile(ctx context.Context, req *model.ProfileCreatedRequest) (int, error)
	UpdateProfile(ctx context.Context, req *model.ProfileUpdatedRequest) error
	DeleteProfile(ctx context.Context, id int) error
}

type HealthUseCase interface {
	Ping(ctx context.Context) model.PingResult
	HealthCheck(ctx context.Context) error
}

package usecase

import (
	"context"
	"courier-service/internal/model"
	"errors"
)

type ProfileUseCase struct {
	repository ProfileRepository
}

func NewProfileUseCase(repository ProfileRepository) *ProfileUseCase {
	return &ProfileUseCase{repository: repository}
}

func (u *ProfileUseCase) GetProfile(ctx context.Context, id int) (*model.Profile, error) {
	profileDB, err := u.repository.GetOneById(ctx, id)
	if err != nil {
		return nil, err
	}

	profile := &model.Profile{
		ID:     profileDB.ID,
		Name:   profileDB.Name,
		Phone:  profileDB.Phone,
		Status: profileDB.Status,
	}

	return profile, nil
}

func (u *ProfileUseCase) GetAllProfiles(ctx context.Context) ([]model.Profile, error) {
	profilesDB, err := u.repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	profiles := []model.Profile{}
	for _, profileDB := range profilesDB {
		profile := model.Profile{
			ID:     profileDB.ID,
			Name:   profileDB.Name,
			Phone:  profileDB.Phone,
			Status: profileDB.Status,
		}
		profiles = append(profiles, profile)
	}

	if profiles == nil {
		profiles = []model.Profile{}
	}

	return profiles, nil
}

func (u *ProfileUseCase) CreateProfile(ctx context.Context, req *model.ProfileCreatedRequest) (int, error) {
	if req.Name == "" || req.Phone == "" || req.Status == "" {
		return 0, ErrMissingRequiredFields
	}

	profileDB := &model.ProfileDB{
		Name:   req.Name,
		Phone:  req.Phone,
		Status: req.Status,
	}

	id, err := u.repository.Create(ctx, profileDB)
	if err != nil {
		if errors.Is(err, ErrPhoneExists) {
			return 0, ErrPhoneExists
		}
		return 0, err
	}

	return id, nil
}

func (u *ProfileUseCase) UpdateProfile(ctx context.Context, req *model.ProfileUpdatedRequest) error {
	if req.Name == "" || req.Phone == "" || req.Status == "" || req.ID == 0 {
		return ErrMissingRequiredFields
	}

	profileDB := &model.ProfileDB{
		ID:     req.ID,
		Name:   req.Name,
		Phone:  req.Phone,
		Status: req.Status,
	}

	err := u.repository.Update(ctx, profileDB)
	if err != nil {
		return err
	}

	return nil
}

func (u *ProfileUseCase) DeleteProfile(ctx context.Context, id int) error {
	err := u.repository.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

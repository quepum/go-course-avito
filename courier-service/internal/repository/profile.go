package repository

import (
	"context"
	"courier-service/internal/model"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepository struct {
	pool *pgxpool.Pool
}

func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

func (r *ProfileRepository) GetOneById(ctx context.Context, id int) (*model.ProfileDB, error) {
	var profile model.ProfileDB
	err := r.pool.QueryRow(ctx, "SELECT id, name, phone, status FROM couriers WHERE id=$1", id).
		Scan(&profile.ID, &profile.Name, &profile.Phone, &profile.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &profile, nil
}

func (r *ProfileRepository) GetAll(ctx context.Context) ([]model.ProfileDB, error) {
	var profiles []model.ProfileDB
	rows, err := r.pool.Query(ctx, "SELECT id, name, phone, status FROM couriers")
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var profile model.ProfileDB
		err := rows.Scan(&profile.ID, &profile.Name, &profile.Phone, &profile.Status)
		if err != nil {
			return nil, fmt.Errorf("reading data error: %w", err)
		}
		profiles = append(profiles, profile)
	}

	if profiles == nil {
		profiles = []model.ProfileDB{}
	}

	return profiles, nil
}

func (r *ProfileRepository) Create(ctx context.Context, profile *model.ProfileDB) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, "INSERT INTO couriers (name, phone, status) VALUES ($1,$2,$3) RETURNING id",
		profile.Name, profile.Phone, profile.Status).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return 0, ErrPhoneExists
		}
		return 0, fmt.Errorf("database error: %w", err)
	}

	return id, nil
}

func (r *ProfileRepository) Update(ctx context.Context, profile *model.ProfileDB) error {
	result, err := r.pool.Exec(ctx, "UPDATE couriers SET name=$1, phone=$2, status=$3 WHERE id=$4",
		profile.Name, profile.Phone, profile.Status, profile.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return ErrPhoneExists
		}
		return fmt.Errorf("database error: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrProfileNotFound
	}

	return nil
}

func (r *ProfileRepository) Delete(ctx context.Context, id int) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM couriers WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrProfileNotFound
	}

	return nil
}

func (r *ProfileRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

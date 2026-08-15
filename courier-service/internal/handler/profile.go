package handler

import (
	"courier-service/internal/model"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ProfileController struct {
	useCase ProfileUseCase
}

func NewProfileController(useCase ProfileUseCase) *ProfileController {
	return &ProfileController{useCase: useCase}
}

func (c *ProfileController) GetOne(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid profile ID"}`, http.StatusBadRequest)
		return
	}

	profile, err := c.useCase.GetProfile(r.Context(), id)

	if err != nil {
		switch {
		case errors.Is(err, ErrProfileNotFound):
			http.Error(w, `{"error": "Profile not found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

func (c *ProfileController) GetAll(w http.ResponseWriter, r *http.Request) {
	profiles, err := c.useCase.GetAllProfiles(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profiles)
}

func (c *ProfileController) Create(w http.ResponseWriter, r *http.Request) {
	var request model.ProfileCreatedRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	id, err := c.useCase.CreateProfile(r.Context(), &request)
	if err != nil {
		switch {
		case errors.Is(err, ErrPhoneExists):
			http.Error(w, `{"error": "Profile with such phone already exists"}`, http.StatusConflict)
		case errors.Is(err, ErrMissingRequiredFields):
			http.Error(w, `{"error": "Missing required field"}`, http.StatusBadRequest)
		default:
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		}

		return
	}

	response := map[string]interface{}{
		"id":      id,
		"message": "Profile created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (c *ProfileController) Update(w http.ResponseWriter, r *http.Request) {
	var request model.ProfileUpdatedRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	err := c.useCase.UpdateProfile(r.Context(), &request)
	if err != nil {
		switch {
		case errors.Is(err, ErrPhoneExists):
			http.Error(w, `{"error": "Profile with such phone already exists"}`, http.StatusConflict)
		case errors.Is(err, ErrProfileNotFound):
			http.Error(w, `{"error": "Profile not found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{
		"message": "Profile updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (c *ProfileController) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid profile ID"}`, http.StatusBadRequest)
		return
	}

	err = c.useCase.DeleteProfile(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrProfileNotFound):
			http.Error(w, `{"error": "Profile not found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{"message": "Profile deleted successfully"}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

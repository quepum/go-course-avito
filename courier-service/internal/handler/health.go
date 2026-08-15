package handler

import (
	"encoding/json"
	"net/http"
)

type HealthController struct {
	useCase HealthUseCase
}

func NewHealthController(useCase HealthUseCase) *HealthController {
	return &HealthController{useCase: useCase}
}

func (u *HealthController) Ping(w http.ResponseWriter, r *http.Request) {
	result := u.useCase.Ping(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (u *HealthController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := u.useCase.HealthCheck(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

package model

type Profile struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

type ProfileCreatedRequest struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

type ProfileUpdatedRequest struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

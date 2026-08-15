package model

type ProfileDB struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

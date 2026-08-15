package model

type PingStatus string

var (
	PingStatusUp   PingStatus = "up"
	PingStatusDown PingStatus = "down"
)

type PingResult struct {
	Status PingStatus `json:"status"`
}

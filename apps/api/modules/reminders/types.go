package reminders

import "time"

type RemindResponse struct {
	Status     string    `json:"status"`
	RemindedAt time.Time `json:"reminded_at"`
}

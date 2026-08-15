package reminders

import "time"

// RemindResponse is returned when a reminder is sent, carrying the timestamp
// the reminder was recorded at.
type RemindResponse struct {
	Status     string    `json:"status"`
	RemindedAt time.Time `json:"reminded_at"`
}

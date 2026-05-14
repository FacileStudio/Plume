package reminders

import (
	"testing"
	"time"
)

func TestShouldRemind(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	ago := func(d time.Duration) time.Time { return now.Add(-d) }
	ptr := func(t time.Time) *time.Time { return &t }

	cases := []struct {
		name         string
		createdAt    time.Time
		lastReminded *time.Time
		intervalDays int
		want         bool
	}{
		{
			name:         "disabled when interval is zero",
			createdAt:    ago(10 * 24 * time.Hour),
			lastReminded: nil,
			intervalDays: 0,
			want:         false,
		},
		{
			name:         "skipped when signer is younger than 24h",
			createdAt:    ago(6 * time.Hour),
			lastReminded: nil,
			intervalDays: 3,
			want:         false,
		},
		{
			name:         "fires when never reminded and older than 24h",
			createdAt:    ago(2 * 24 * time.Hour),
			lastReminded: nil,
			intervalDays: 3,
			want:         true,
		},
		{
			name:         "skipped when last reminder is within interval",
			createdAt:    ago(10 * 24 * time.Hour),
			lastReminded: ptr(ago(1 * 24 * time.Hour)),
			intervalDays: 3,
			want:         false,
		},
		{
			name:         "fires when last reminder is older than interval",
			createdAt:    ago(10 * 24 * time.Hour),
			lastReminded: ptr(ago(4 * 24 * time.Hour)),
			intervalDays: 3,
			want:         true,
		},
		{
			name:         "fires exactly at interval boundary",
			createdAt:    ago(10 * 24 * time.Hour),
			lastReminded: ptr(ago(3 * 24 * time.Hour)),
			intervalDays: 3,
			want:         true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldRemind(tc.createdAt, tc.lastReminded, tc.intervalDays, now)
			if got != tc.want {
				t.Fatalf("shouldRemind() = %v, want %v", got, tc.want)
			}
		})
	}
}

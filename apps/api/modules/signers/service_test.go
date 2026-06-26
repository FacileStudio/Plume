package signers

import (
	"testing"

	"api/schemas"
)

func TestIsSignersTurn(t *testing.T) {
	sig := func(id int64, role, status string, order int) schemas.Signer {
		return schemas.Signer{ID: id, Role: role, Status: status, OrderNum: order}
	}

	cases := []struct {
		name       string
		sequential bool
		current    schemas.Signer
		all        []schemas.Signer
		want       bool
	}{
		{
			name:       "non-sequential always allowed",
			sequential: false,
			current:    sig(2, "signer", "pending", 2),
			all:        []schemas.Signer{sig(1, "signer", "pending", 1), sig(2, "signer", "pending", 2)},
			want:       true,
		},
		{
			name:       "first in order may sign",
			sequential: true,
			current:    sig(1, "signer", "pending", 1),
			all:        []schemas.Signer{sig(1, "signer", "pending", 1), sig(2, "signer", "pending", 2)},
			want:       true,
		},
		{
			name:       "blocked while earlier signer still pending",
			sequential: true,
			current:    sig(2, "signer", "pending", 2),
			all:        []schemas.Signer{sig(1, "signer", "pending", 1), sig(2, "signer", "pending", 2)},
			want:       false,
		},
		{
			name:       "allowed once earlier signer signed",
			sequential: true,
			current:    sig(2, "signer", "pending", 2),
			all:        []schemas.Signer{sig(1, "signer", "signed", 1), sig(2, "signer", "pending", 2)},
			want:       true,
		},
		{
			name:       "earlier viewer does not block",
			sequential: true,
			current:    sig(2, "signer", "pending", 2),
			all:        []schemas.Signer{sig(1, "viewer", "pending", 1), sig(2, "signer", "pending", 2)},
			want:       true,
		},
		{
			name:       "ties at same order may sign in parallel",
			sequential: true,
			current:    sig(2, "signer", "pending", 1),
			all:        []schemas.Signer{sig(1, "signer", "pending", 1), sig(2, "signer", "pending", 1)},
			want:       true,
		},
		{
			name:       "current viewer never gated",
			sequential: true,
			current:    sig(3, "viewer", "pending", 5),
			all:        []schemas.Signer{sig(1, "signer", "pending", 1), sig(3, "viewer", "pending", 5)},
			want:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSignersTurn(tc.sequential, tc.current, tc.all); got != tc.want {
				t.Fatalf("isSignersTurn() = %v, want %v", got, tc.want)
			}
		})
	}
}

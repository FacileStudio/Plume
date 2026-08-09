package auth

import "testing"

func TestIsEmailVerified(t *testing.T) {
	cases := []struct {
		claim any
		want  bool
	}{
		{nil, true},
		{true, true},
		{false, false},
		{"true", true},
		{"false", false},
		{"FALSE", false},
	}
	for _, c := range cases {
		if got := isEmailVerified(c.claim); got != c.want {
			t.Errorf("isEmailVerified(%#v) = %v, want %v", c.claim, got, c.want)
		}
	}
}

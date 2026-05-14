package verify

import "testing"

func TestMaskEmail(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"alice@example.com", "a****@e**********"},
		{"a@b.co", "*@b***"},
		{"", "***"},
		{"noatsign", "***"},
		{"@nolocal.com", "***"},
		{"nolocal@", "***"},
		{"x@y", "*@*"},
	}
	for _, tc := range cases {
		got := maskEmail(tc.in)
		if got != tc.want {
			t.Errorf("maskEmail(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

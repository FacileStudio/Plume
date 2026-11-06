package env

import "testing"

func TestLoadNuage(t *testing.T) {
	oidc := &OIDCConfig{Issuer: "https://sso.test"}
	cases := []struct {
		name     string
		url      string
		audience string
		oidc     *OIDCConfig
		wantNil  bool
		wantErr  bool
	}{
		{"unset", "", "", nil, true, false},
		{"complete", "https://nuage.test/", "nuage-client", oidc, false, false},
		{"url only", "https://nuage.test", "", oidc, false, true},
		{"audience only", "", "nuage-client", oidc, false, true},
		{"bad scheme", "nuage.test", "nuage-client", oidc, false, true},
		{"without sso", "https://nuage.test", "nuage-client", nil, false, true},
	}
	for _, c := range cases {
		t.Setenv("NUAGE_URL", c.url)
		t.Setenv("NUAGE_AUDIENCE", c.audience)
		cfg, err := loadNuage(c.oidc)
		if (err != nil) != c.wantErr || (cfg == nil) != (c.wantNil || c.wantErr) {
			t.Errorf("%s: cfg=%v err=%v", c.name, cfg, err)
		}
	}
}

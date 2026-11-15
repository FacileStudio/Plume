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

func TestLoadReadsSMTPAllowPrivateHosts(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("APP_ENV", "development")
	for value, want := range map[string]bool{"": false, "false": false, "true": true} {
		t.Setenv("SMTP_ALLOW_PRIVATE_HOSTS", value)
		cfg, err := Load()
		if err != nil || cfg.SMTPAllowPrivateHosts != want {
			t.Errorf("SMTP_ALLOW_PRIVATE_HOSTS=%q: allow=%v err=%v, want %v", value, cfg.SMTPAllowPrivateHosts, err, want)
		}
	}
	t.Setenv("SMTP_ALLOW_PRIVATE_HOSTS", "maybe")
	if _, err := Load(); err == nil {
		t.Error("a non-boolean SMTP_ALLOW_PRIVATE_HOSTS must fail startup")
	}
}

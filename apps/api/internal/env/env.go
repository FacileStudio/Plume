package env

import (
	"github.com/FacileStudio/porte"

	"fmt"
	"strings"

	troncenv "github.com/FacileStudio/tronc/env"
)

// OIDCConfig holds the OIDC provider settings loaded from the environment.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	SuccessURL   string
}

// Config holds the application's environment-derived configuration.
type Config struct {
	troncenv.Core
	Domain    string
	UploadDir string
	OIDC      *OIDCConfig
	SSOOnly   bool
}

// Load reads and validates the application configuration from the
// environment, including optional OIDC settings.
func Load() (Config, error) {
	core, err := troncenv.LoadCore()
	if err != nil {
		return Config{}, err
	}
	if core.Port < 1 || core.Port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a valid TCP port")
	}
	if err := validateLogLevel(core.LogLevel); err != nil {
		return Config{}, err
	}

	ssoOnly, err := troncenv.Bool("SSO_ONLY", false)
	if err != nil {
		return Config{}, err
	}

	env := Config{
		Core:      core,
		Domain:    troncenv.String("DOMAIN", "http://localhost:5173"),
		UploadDir: troncenv.String("UPLOAD_DIR", "/data/uploads"),
		SSOOnly:   ssoOnly,
	}

	if issuer := troncenv.String("OIDC_ISSUER", ""); issuer != "" {
		clientID := troncenv.String("OIDC_CLIENT_ID", "")
		clientSecret := troncenv.String("OIDC_CLIENT_SECRET", "")
		redirectURL := troncenv.String("OIDC_REDIRECT_URL", "")
		if clientID == "" || clientSecret == "" || redirectURL == "" {
			return Config{}, fmt.Errorf("OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL are required when OIDC_ISSUER is set")
		}
		env.OIDC = &OIDCConfig{
			Issuer:       issuer,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			SuccessURL:   troncenv.String("OIDC_SUCCESS_URL", env.Domain),
		}
	}

	return env, nil
}

func validateLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error")
	}
}

// Porte is the one configuration porte's session manager, OIDC kit and local
// login are all built from. They share it because porte refuses at boot a kit
// whose config disagrees with its manager's — a mismatch would otherwise
// change silently whether the session cookie is Secure.
//
// AcceptLegacyCookie is on even though this app has always been bearer-only:
// it costs nothing when no such cookie exists, and porte now sets one, so the
// setting describes the transport rather than the migration.
func (c Config) Porte() porte.Config {
	cfg := porte.Config{SSOOnly: c.SSOOnly, AcceptLegacyCookie: true}
	if c.OIDC == nil {
		return cfg
	}
	cfg.Issuer = c.OIDC.Issuer
	cfg.ClientID = c.OIDC.ClientID
	cfg.ClientSecret = c.OIDC.ClientSecret
	cfg.RedirectURL = c.OIDC.RedirectURL
	cfg.SuccessURL = c.OIDC.SuccessURL
	return cfg
}

// IssuerForMigration is the issuer the identity backfill keys on, or empty
// when SSO is not configured. It exists so the migration cannot be handed a
// placeholder: an identity row written under the wrong provider matches
// nothing and degrades to the email fallback in silence.
func (c Config) IssuerForMigration() string {
	if c.OIDC == nil {
		return ""
	}
	return c.OIDC.Issuer
}

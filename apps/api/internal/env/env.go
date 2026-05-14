package env

import (
	"fmt"
	"strconv"
	"strings"
)

type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	SuccessURL   string
}

type Config struct {
	DatabaseURL string
	Port        string
	Domain      string
	LogLevel    string
	OIDC        *OIDCConfig
	SSOOnly     bool
}

func Load() (Config, error) {
	env := Config{
		DatabaseURL: valueOrDefault("DATABASE_URL", "postgres://postgres:postgres@db:5432/plume?sslmode=disable"),
		Port:        valueOrDefault("PORT", "4000"),
		Domain:      valueOrDefault("DOMAIN", "http://localhost:5173"),
		LogLevel:    valueOrDefault("LOG_LEVEL", "info"),
	}

	port, err := strconv.Atoi(env.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a valid TCP port")
	}
	if err := validateLogLevel(env.LogLevel); err != nil {
		return Config{}, err
	}

	env.SSOOnly = strings.ToLower(envGet("SSO_ONLY")) == "true"

	if issuer := envGet("OIDC_ISSUER"); issuer != "" {
		clientID := envGet("OIDC_CLIENT_ID")
		clientSecret := envGet("OIDC_CLIENT_SECRET")
		redirectURL := envGet("OIDC_REDIRECT_URL")
		if clientID == "" || clientSecret == "" || redirectURL == "" {
			return Config{}, fmt.Errorf("OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL are required when OIDC_ISSUER is set")
		}
		successURL := envGet("OIDC_SUCCESS_URL")
		if successURL == "" {
			successURL = env.Domain
		}
		env.OIDC = &OIDCConfig{
			Issuer:       issuer,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			SuccessURL:   successURL,
		}
	}

	return env, nil
}

func valueOrDefault(key string, fallback string) string {
	if value := envGet(key); value != "" {
		return value
	}
	return fallback
}

func validateLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error")
	}
}

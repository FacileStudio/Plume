package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"api/internal/env"
	"api/internal/errors"
	"api/internal/httpjson"
	"api/internal/oidcavatar"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	oidcStateCookie = "oidc_state"
	oidcStateTTL    = 10 * time.Minute
)

type pendingCode struct {
	UserID    string
	Token     string
	ExpiresAt time.Time
}

type oidcHandler struct {
	provider   *gooidc.Provider
	verifier   *gooidc.IDTokenVerifier
	oauth2Cfg  oauth2.Config
	service    *Service
	successURL string
	codes      sync.Map
}

func newOIDCHandler(ctx context.Context, cfg *env.OIDCConfig, service *Service) (*oidcHandler, error) {
	provider, err := gooidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}
	oauth2Cfg := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{gooidc.ScopeOpenID, "email", "profile"},
	}
	verifier := provider.Verifier(&gooidc.Config{ClientID: cfg.ClientID})
	return &oidcHandler{
		provider:   provider,
		verifier:   verifier,
		oauth2Cfg:  oauth2Cfg,
		service:    service,
		successURL: cfg.SuccessURL,
	}, nil
}

func (h *oidcHandler) login(w http.ResponseWriter, r *http.Request) {
	state, err := randomState()
	if err != nil {
		httpjson.WriteError(w, errors.Internal("failed to generate state", err))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oidcStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   int(oidcStateTTL.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, h.oauth2Cfg.AuthCodeURL(state), http.StatusFound)
}

func (h *oidcHandler) callback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(oidcStateCookie)
	if err != nil || subtle.ConstantTimeCompare([]byte(stateCookie.Value), []byte(r.URL.Query().Get("state"))) != 1 {
		httpjson.WriteError(w, errors.Invalid("invalid oauth2 state"))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oidcStateCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})

	oauth2Token, err := h.oauth2Cfg.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		httpjson.WriteError(w, errors.Internal("failed to exchange code", err))
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		httpjson.WriteError(w, errors.Internal("missing id_token in response", nil))
		return
	}

	idToken, err := h.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		httpjson.WriteError(w, errors.Unauthorized("invalid id_token"))
		return
	}

	var claims struct {
		Email             string `json:"email"`
		EmailVerified     any    `json:"email_verified"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
		GivenName         string `json:"given_name"`
		FamilyName        string `json:"family_name"`
		Picture           string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		httpjson.WriteError(w, errors.Internal("failed to parse claims", err))
		return
	}
	if claims.Email == "" {
		httpjson.WriteError(w, errors.Invalid("OIDC provider did not return an email"))
		return
	}
	if !isEmailVerified(claims.EmailVerified) {
		httpjson.WriteError(w, errors.Invalid("email not verified by your identity provider"))
		return
	}

	profile := oidcavatar.Profile{
		Name:             claims.Name,
		PreferredUsername: claims.PreferredUsername,
		GivenName:        claims.GivenName,
		FamilyName:       claims.FamilyName,
		Picture:          claims.Picture,
	}
	userID, token, err := h.service.upsertOIDCUser(r.Context(), claims.Email, profile)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}

	code, err := randomState()
	if err != nil {
		httpjson.WriteError(w, errors.Internal("failed to generate auth code", err))
		return
	}
	h.codes.Store(code, pendingCode{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(60 * time.Second),
	})

	dest, _ := url.Parse(h.successURL)
	q := dest.Query()
	q.Set("code", code)
	dest.RawQuery = q.Encode()
	http.Redirect(w, r, dest.String(), http.StatusFound)
}

func (h *oidcHandler) exchange(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := httpjson.DecodeJSON(w, r, &req); err != nil {
		httpjson.WriteError(w, err)
		return
	}

	val, ok := h.codes.LoadAndDelete(req.Code)
	if !ok {
		httpjson.WriteError(w, errors.Unauthorized("invalid or expired login code"))
		return
	}

	pending := val.(pendingCode)
	if time.Now().After(pending.ExpiresAt) {
		httpjson.WriteError(w, errors.Unauthorized("invalid or expired login code"))
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, AuthResponse{
		UserID: pending.UserID,
		Token:  pending.Token,
	})
}

func isEmailVerified(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return strings.EqualFold(val, "true")
	default:
		return false
	}
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

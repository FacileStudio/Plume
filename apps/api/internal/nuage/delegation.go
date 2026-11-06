package nuage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/FacileStudio/porte"
)

const (
	exchangeGrant   = "urn:ietf:params:oauth:grant-type:token-exchange"
	accessTokenType = "urn:ietf:params:oauth:token-type:access_token"
	writeScope      = "nuage.write"
	tokenMargin     = 30 * time.Second
	maxBody         = 1 << 20
)

// ErrNotConnected means the user has no usable Registre session for Plume: no
// stored refresh token, or one Registre no longer accepts. Retrying cannot
// help until the user signs in again.
var ErrNotConnected = errors.New("user has no Registre session Plume can act with")

// Identities is the slice of porte's identity store the delegation needs.
type Identities interface {
	ListByUser(ctx context.Context, userID int64) ([]porte.StoredIdentity, error)
	Save(ctx context.Context, identity porte.StoredIdentity) error
}

// DelegationConfig names the Registre client Plume authenticates as and the
// audience Nuage verifies, which is Nuage's own Registre client id.
type DelegationConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	Audience     string
}

// Delegation turns a Plume user into a five-minute Nuage token: it refreshes
// the user's stored Registre access token when needed and trades it at
// Registre's token endpoint (RFC 8693) for a token whose act.sub is Plume.
type Delegation struct {
	cfg    DelegationConfig
	ids    Identities
	client *http.Client
	now    func() time.Time

	mu         sync.Mutex
	endpointMu sync.Mutex
	endpoint   string
}

// NewDelegation builds a Delegation over the identity store.
func NewDelegation(cfg DelegationConfig, ids Identities) *Delegation {
	return &Delegation{cfg: cfg, ids: ids, client: &http.Client{Timeout: requestTimeout}, now: time.Now}
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
}

// Token returns a Nuage token for userID, valid for one app-folder write.
func (d *Delegation) Token(ctx context.Context, userID int64) (string, error) {
	for _, force := range []bool{false, true} {
		subject, err := d.subject(ctx, userID, force)
		if err != nil {
			return "", err
		}
		out, err := d.post(ctx, url.Values{
			"grant_type":         {exchangeGrant},
			"subject_token":      {subject},
			"subject_token_type": {accessTokenType},
			"audience":           {d.cfg.Audience},
			"scope":              {writeScope},
		})
		if err == nil {
			return out.AccessToken, nil
		}
		if out.Error != "invalid_grant" {
			return "", err
		}
	}
	return "", ErrNotConnected
}

func (d *Delegation) subject(ctx context.Context, userID int64, force bool) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	identity, err := d.identity(ctx, userID)
	if err != nil {
		return "", err
	}
	tokens := identity.Tokens
	if !force && tokens.AccessToken != "" && d.now().Add(tokenMargin).Before(tokens.Expiry) {
		return tokens.AccessToken, nil
	}
	if tokens.RefreshToken == "" {
		return "", ErrNotConnected
	}
	return d.refresh(ctx, identity)
}

func (d *Delegation) refresh(ctx context.Context, identity porte.StoredIdentity) (string, error) {
	out, err := d.post(ctx, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {identity.Tokens.RefreshToken}})
	if err != nil {
		if out.Error == "invalid_grant" {
			return "", ErrNotConnected
		}
		return "", err
	}
	next := out.RefreshToken
	if next == "" {
		next = identity.Tokens.RefreshToken
	}
	identity.Tokens = porte.TokenSet{
		AccessToken:  out.AccessToken,
		RefreshToken: next,
		Expiry:       d.now().Add(time.Duration(out.ExpiresIn) * time.Second),
	}
	if err := d.ids.Save(ctx, identity); err != nil {
		return "", fmt.Errorf("save refreshed token: %w", err)
	}
	return out.AccessToken, nil
}

func (d *Delegation) identity(ctx context.Context, userID int64) (porte.StoredIdentity, error) {
	stored, err := d.ids.ListByUser(ctx, userID)
	if err != nil {
		return porte.StoredIdentity{}, fmt.Errorf("read identities: %w", err)
	}
	for _, identity := range stored {
		if identity.Provider == d.cfg.Issuer && identity.PasswordHash == "" {
			return identity, nil
		}
	}
	return porte.StoredIdentity{}, ErrNotConnected
}

func (d *Delegation) tokenEndpoint(ctx context.Context) (string, error) {
	d.endpointMu.Lock()
	defer d.endpointMu.Unlock()
	if d.endpoint != "" {
		return d.endpoint, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(d.cfg.Issuer, "/")+"/.well-known/openid-configuration", nil)
	if err != nil {
		return "", err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("registre discovery: %w", err)
	}
	defer resp.Body.Close()
	var doc struct {
		TokenEndpoint string `json:"token_endpoint"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBody)).Decode(&doc); err != nil || doc.TokenEndpoint == "" {
		return "", errors.New("registre discovery returned no token endpoint")
	}
	d.endpoint = doc.TokenEndpoint
	return d.endpoint, nil
}

func (d *Delegation) post(ctx context.Context, form url.Values) (tokenResponse, error) {
	endpoint, err := d.tokenEndpoint(ctx)
	if err != nil {
		return tokenResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return tokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(url.QueryEscape(d.cfg.ClientID), url.QueryEscape(d.cfg.ClientSecret))
	resp, err := d.client.Do(req)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("registre token call: %w", err)
	}
	defer resp.Body.Close()
	var out tokenResponse
	_ = json.NewDecoder(io.LimitReader(resp.Body, maxBody)).Decode(&out)
	if resp.StatusCode != http.StatusOK || out.AccessToken == "" {
		return out, &StatusError{Service: "registre", Status: resp.StatusCode, Body: out.Error}
	}
	return out, nil
}

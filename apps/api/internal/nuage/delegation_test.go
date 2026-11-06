package nuage

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/porte"
)

type fakeIdentities struct {
	mu        sync.Mutex
	stored    []porte.StoredIdentity
	failSaves int
	saves     int
}

func (f *fakeIdentities) ListByUser(context.Context, int64) ([]porte.StoredIdentity, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]porte.StoredIdentity(nil), f.stored...), nil
}

func (f *fakeIdentities) Save(_ context.Context, identity porte.StoredIdentity) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.saves++
	if f.saves <= f.failSaves {
		return errors.New("database down")
	}
	f.stored = []porte.StoredIdentity{identity}
	return nil
}

type fakeRegistre struct {
	mu        sync.Mutex
	forms     []url.Values
	user      string
	pass      string
	exchanges []string
	reject    map[string]bool
	revoked   map[string]bool
}

func (f *fakeRegistre) serve(t *testing.T) (*httptest.Server, DelegationConfig) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			_ = json.NewEncoder(w).Encode(map[string]string{"token_endpoint": srv.URL + "/token"})
			return
		}
		_ = r.ParseForm()
		f.mu.Lock()
		defer f.mu.Unlock()
		f.user, f.pass, _ = r.BasicAuth()
		f.forms = append(f.forms, r.PostForm)
		grant := r.PostForm.Get("grant_type")
		if f.reject[grant] {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
			return
		}
		if grant == "refresh_token" {
			_, _ = w.Write([]byte(`{"access_token":"fresh","refresh_token":"rotated","expires_in":900}`))
			return
		}
		f.exchanges = append(f.exchanges, r.PostForm.Get("subject_token"))
		if f.revoked[r.PostForm.Get("subject_token")] {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"delegated","expires_in":300}`))
	}))
	t.Cleanup(srv.Close)
	return srv, DelegationConfig{Issuer: srv.URL, ClientID: "plume", ClientSecret: "s3cret", Audience: "nuage-client"}
}

func identityFor(issuer string, tokens porte.TokenSet) porte.StoredIdentity {
	return porte.StoredIdentity{UserID: 1, Provider: issuer, Subject: "sub", Tokens: tokens}
}

func TestTokenExchangesTheStoredAccessTokenForNuageWrite(t *testing.T) {
	reg := &fakeRegistre{}
	_, cfg := reg.serve(t)
	ids := &fakeIdentities{stored: []porte.StoredIdentity{identityFor(cfg.Issuer, porte.TokenSet{
		AccessToken: "stored", RefreshToken: "r", Expiry: time.Now().Add(10 * time.Minute),
	})}}

	token, err := NewDelegation(cfg, ids, quietLogger()).Token(context.Background(), 1)
	if err != nil || token != "delegated" {
		t.Fatalf("token=%q err=%v", token, err)
	}
	form := reg.forms[0]
	if form.Get("grant_type") != exchangeGrant || form.Get("subject_token") != "stored" ||
		form.Get("audience") != "nuage-client" || form.Get("scope") != "nuage.write" || form.Get("subject_token_type") != accessTokenType {
		t.Fatalf("bad exchange form %v", form)
	}
	if reg.user != "plume" || reg.pass != "s3cret" || len(reg.forms) != 1 {
		t.Fatalf("auth %q/%q forms=%d", reg.user, reg.pass, len(reg.forms))
	}
}

func TestTokenRefreshesAnExpiredAccessTokenAndSavesTheRotation(t *testing.T) {
	reg := &fakeRegistre{}
	_, cfg := reg.serve(t)
	ids := &fakeIdentities{stored: []porte.StoredIdentity{identityFor(cfg.Issuer, porte.TokenSet{
		AccessToken: "old", RefreshToken: "r", Expiry: time.Now().Add(-time.Minute),
	})}}

	if _, err := NewDelegation(cfg, ids, quietLogger()).Token(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if reg.exchanges[0] != "fresh" || ids.stored[0].Tokens.RefreshToken != "rotated" || ids.stored[0].Tokens.AccessToken != "fresh" {
		t.Fatalf("exchanged %v saved %+v", reg.exchanges, ids.stored[0].Tokens)
	}
}

func TestTokenRetriesOnceWithAForcedRefreshWhenRegistreRejectsTheSubject(t *testing.T) {
	reg := &fakeRegistre{revoked: map[string]bool{"revoked": true}}
	_, cfg := reg.serve(t)
	ids := &fakeIdentities{stored: []porte.StoredIdentity{identityFor(cfg.Issuer, porte.TokenSet{
		AccessToken: "revoked", RefreshToken: "r", Expiry: time.Now().Add(10 * time.Minute),
	})}}

	token, err := NewDelegation(cfg, ids, quietLogger()).Token(context.Background(), 1)
	if err != nil || token != "delegated" || len(reg.exchanges) != 2 || reg.exchanges[1] != "fresh" {
		t.Fatalf("token=%q err=%v exchanges=%v", token, err, reg.exchanges)
	}
}

func TestTokenIsNotConnectedWhenTheRefreshTokenIsDead(t *testing.T) {
	reg := &fakeRegistre{reject: map[string]bool{"refresh_token": true}}
	_, cfg := reg.serve(t)
	ids := &fakeIdentities{stored: []porte.StoredIdentity{identityFor(cfg.Issuer, porte.TokenSet{RefreshToken: "dead"})}}

	_, err := NewDelegation(cfg, ids, quietLogger()).Token(context.Background(), 1)
	if !errors.Is(err, ErrNotConnected) || len(reg.exchanges) != 0 {
		t.Fatalf("err=%v exchanges=%v", err, reg.exchanges)
	}
}

func TestTokenIsNotConnectedForPasswordOnlyOrTokenlessUsers(t *testing.T) {
	reg := &fakeRegistre{}
	_, cfg := reg.serve(t)
	cases := map[string][]porte.StoredIdentity{
		"no identity":   nil,
		"password only": {{UserID: 1, Provider: "local", PasswordHash: "h"}},
		"no refresh":    {identityFor(cfg.Issuer, porte.TokenSet{})},
	}
	for name, stored := range cases {
		_, err := NewDelegation(cfg, &fakeIdentities{stored: stored}, quietLogger()).Token(context.Background(), 1)
		if !errors.Is(err, ErrNotConnected) {
			t.Errorf("%s: err=%v", name, err)
		}
	}
	if len(reg.forms) != 0 {
		t.Errorf("Registre was called %d times", len(reg.forms))
	}
}

func TestConcurrentTokensForOneUserRefreshOnce(t *testing.T) {
	reg := &fakeRegistre{}
	_, cfg := reg.serve(t)
	ids := &fakeIdentities{stored: []porte.StoredIdentity{identityFor(cfg.Issuer, porte.TokenSet{
		AccessToken: "old", RefreshToken: "r", Expiry: time.Now().Add(-time.Minute),
	})}}
	d := NewDelegation(cfg, ids, quietLogger())

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := d.Token(context.Background(), 1); err != nil {
				t.Errorf("token: %v", err)
			}
		}()
	}
	wg.Wait()
	refreshes := 0
	for _, form := range reg.forms {
		if form.Get("grant_type") == "refresh_token" {
			refreshes++
		}
	}
	if refreshes != 1 {
		t.Fatalf("refresh grants = %d, want 1", refreshes)
	}
}

func TestSaveOfARotatedTokenIsRetried(t *testing.T) {
	reg := &fakeRegistre{}
	_, cfg := reg.serve(t)
	ids := &fakeIdentities{failSaves: 2, stored: []porte.StoredIdentity{identityFor(cfg.Issuer, porte.TokenSet{
		AccessToken: "old", RefreshToken: "r", Expiry: time.Now().Add(-time.Minute),
	})}}
	d := NewDelegation(cfg, ids, quietLogger())
	d.saveBackoff = time.Millisecond

	if _, err := d.Token(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if ids.saves != 3 || ids.stored[0].Tokens.RefreshToken != "rotated" {
		t.Fatalf("saves=%d stored=%+v", ids.saves, ids.stored[0].Tokens)
	}
}

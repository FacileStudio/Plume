package auth

import (
	"context"
	"net/http"
	"testing"
)

const (
	firstPassword = "correct horse battery"
	nextPassword  = "staple correct horse"
)

// A federated account has no password identity, so PUT /auth/password without
// current_password is an addition rather than a replacement. Nothing is
// rotated, so no token comes back, and the password has to actually work
// afterwards — the only honest proof that the identity was keyed the way
// porte/local looks one up.
func TestAddingAFirstPasswordToAFederatedAccount(t *testing.T) {
	api := newAPI(t)
	_, token := api.federated(t, "iris@facile.studio")

	body := decode(t, api.do(t, http.MethodPut, "/auth/password", token, map[string]string{
		"new_password": firstPassword,
	}), http.StatusOK)
	if _, rotated := body["token"]; rotated {
		t.Fatalf("adding a first password rotated the session: %v", body)
	}

	login := decode(t, api.do(t, http.MethodPost, "/auth/login", "", map[string]string{
		"email": "iris@facile.studio", "password": firstPassword,
	}), http.StatusOK)
	if login["token"] == "" {
		t.Fatalf("the first password does not log in: %v", login)
	}
}

// Changing a password ends the account's other logins, rotates the caller's
// own, and leaves a named API token alone. The rotated token is in the body
// because this app's client holds a bearer in localStorage: without it the
// browser that made the change keeps sending a token porte revoked mid-request.
func TestChangingAPasswordRotatesTheCallerAndEndsTheOtherLogins(t *testing.T) {
	api := newAPI(t)
	userID, caller := api.register(t, "noah@facile.studio", firstPassword)

	other := decode(t, api.do(t, http.MethodPost, "/auth/login", "", map[string]string{
		"email": "noah@facile.studio", "password": firstPassword,
	}), http.StatusOK)["token"].(string)
	apiToken, _, err := api.service.Sessions().Issue(context.Background(), userID, "ci")
	if err != nil {
		t.Fatalf("issue a named token: %v", err)
	}

	body := decode(t, api.do(t, http.MethodPut, "/auth/password", caller, map[string]string{
		"current_password": firstPassword, "new_password": nextPassword,
	}), http.StatusOK)
	rotated, _ := body["token"].(string)
	if rotated == "" || rotated == caller {
		t.Fatalf("the caller's session was not rotated: %v", body)
	}

	if code := api.do(t, http.MethodGet, "/auth/me", rotated, nil).Code; code != http.StatusOK {
		t.Fatalf("the rotated token does not authenticate: %d", code)
	}
	if code := api.do(t, http.MethodGet, "/auth/me", caller, nil).Code; code != http.StatusUnauthorized {
		t.Fatalf("the old token still authenticates: %d", code)
	}
	if code := api.do(t, http.MethodGet, "/auth/me", other, nil).Code; code != http.StatusUnauthorized {
		t.Fatalf("the other browser was not signed out: %d", code)
	}
	if code := api.do(t, http.MethodGet, "/auth/me", apiToken, nil).Code; code != http.StatusOK {
		t.Fatalf("a named API token was revoked by a password change: %d", code)
	}
}

// The wrong current password is 401 and changes nothing.
func TestChangingAPasswordRefusesTheWrongCurrentOne(t *testing.T) {
	api := newAPI(t)
	_, caller := api.register(t, "noah@facile.studio", firstPassword)

	body := decode(t, api.do(t, http.MethodPut, "/auth/password", caller, map[string]string{
		"current_password": "not the password", "new_password": nextPassword,
	}), http.StatusUnauthorized)
	if code := errorCode(body); code != "unauthenticated" {
		t.Fatalf("error code %q, want unauthenticated: %v", code, body)
	}
	if code := api.do(t, http.MethodGet, "/auth/me", caller, nil).Code; code != http.StatusOK {
		t.Fatalf("a refused change ended the caller's session: %d", code)
	}
}

// An account that already has a password and sends only the new one omitted a
// field; that is 400 naming the field, not porte's 409, and the password it has
// is untouched.
func TestANewPasswordAloneIsRefusedWhenOneIsAlreadySet(t *testing.T) {
	api := newAPI(t)
	_, caller := api.register(t, "noah@facile.studio", firstPassword)

	recorder := api.do(t, http.MethodPut, "/auth/password", caller, map[string]string{
		"new_password": nextPassword,
	})
	body := decode(t, recorder, http.StatusBadRequest)
	if code := errorCode(body); code != "invalid_argument" {
		t.Fatalf("error code %q, want invalid_argument: %v", code, body)
	}

	login := api.do(t, http.MethodPost, "/auth/login", "", map[string]string{
		"email": "noah@facile.studio", "password": firstPassword,
	})
	if login.Code != http.StatusOK {
		t.Fatalf("the refused request changed the password: %d", login.Code)
	}
}

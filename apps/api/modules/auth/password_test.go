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

// The wrong current password is 401 and changes nothing. The message is this
// app's, not porte's: the form has no email field, so porte's "invalid email
// or password" describes a field the person is not looking at, and the package
// name is nobody's business but ours.
func TestChangingAPasswordRefusesTheWrongCurrentOne(t *testing.T) {
	api := newAPI(t)
	_, caller := api.register(t, "noah@facile.studio", firstPassword)

	body := decode(t, api.do(t, http.MethodPut, "/auth/password", caller, map[string]string{
		"current_password": "not the password", "new_password": nextPassword,
	}), http.StatusUnauthorized)
	if code := errorCode(body); code != "unauthenticated" {
		t.Fatalf("error code %q, want unauthenticated: %v", code, body)
	}
	if message := errorMessage(body); message != "current password is incorrect" {
		t.Fatalf("message %q, want the app's own wording", message)
	}
	if code := api.do(t, http.MethodGet, "/auth/me", caller, nil).Code; code != http.StatusOK {
		t.Fatalf("a refused change ended the caller's session: %d", code)
	}
}

// Eleven characters is refused and twelve is accepted. The client's minlength
// mirrors this number, and it mirrored 8 for a while, so the floor is pinned
// here rather than left to the caption on the form.
func TestThePasswordFloorIsTwelveCharacters(t *testing.T) {
	api := newAPI(t)
	_, caller := api.register(t, "noah@facile.studio", firstPassword)

	short := api.do(t, http.MethodPut, "/auth/password", caller, map[string]string{
		"current_password": firstPassword, "new_password": "elevenchars",
	})
	if short.Code != http.StatusBadRequest {
		t.Fatalf("an 11-character password got %d, want 400", short.Code)
	}
	long := api.do(t, http.MethodPut, "/auth/password", caller, map[string]string{
		"current_password": firstPassword, "new_password": "twelvechars!",
	})
	if long.Code != http.StatusOK {
		t.Fatalf("a 12-character password got %d, want 200: %s", long.Code, long.Body.String())
	}
}

// A federated account that sends a current password has none to confirm. That
// is 400, and the sentence has to explain the account rather than name the
// library that noticed.
func TestChangingAPasswordOnAnAccountThatHasNone(t *testing.T) {
	api := newAPI(t)
	_, caller := api.federated(t, "iris@facile.studio")

	body := decode(t, api.do(t, http.MethodPut, "/auth/password", caller, map[string]string{
		"current_password": firstPassword, "new_password": nextPassword,
	}), http.StatusBadRequest)
	if message := errorMessage(body); message != "this account has no password to change" {
		t.Fatalf("message %q, want the app's own wording", message)
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

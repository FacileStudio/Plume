package auth

import (
	"net/http"
	"testing"
)

func TestAPITokenAuthenticatesLikeASessionAndNeedsALabel(t *testing.T) {
	api := newAPI(t)
	_, session := api.register(t, "ardoise@facile.studio", firstPassword)

	decode(t, api.do(t, http.MethodPost, "/auth/api-token", session, map[string]string{"label": " "}), http.StatusBadRequest)
	decode(t, api.do(t, http.MethodPost, "/auth/api-token", "", map[string]string{"label": "ardoise"}), http.StatusUnauthorized)

	body := decode(t, api.do(t, http.MethodPost, "/auth/api-token", session, map[string]string{"label": "ardoise"}), http.StatusCreated)
	token, _ := body["token"].(string)
	if token == "" || body["label"] != "ardoise" {
		t.Fatalf("unexpected response %v", body)
	}
	me := decode(t, api.do(t, http.MethodGet, "/auth/me", token, nil), http.StatusOK)
	if me["email"] != "ardoise@facile.studio" {
		t.Fatalf("token does not act as its user: %v", me)
	}
}

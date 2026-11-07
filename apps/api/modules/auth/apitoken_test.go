package auth

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestAPITokenAuthenticatesLikeASessionAndNeedsALabel(t *testing.T) {
	api := newAPI(t)
	_, session := api.federated(t, "ardoise@facile.studio")

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

func TestAPITokenListAndRevokeNeverTouchLoginSessions(t *testing.T) {
	api := newAPI(t)
	_, session := api.federated(t, "owner@facile.studio")
	_, other := api.federated(t, "other@facile.studio")

	created := decode(t, api.do(t, http.MethodPost, "/auth/api-token", session, map[string]string{"label": "ardoise"}), http.StatusCreated)
	token, _ := created["token"].(string)
	id := int64(created["id"].(float64))

	listed := api.do(t, http.MethodGet, "/auth/api-tokens", session, nil)
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), `"ardoise"`) || strings.Contains(listed.Body.String(), token) {
		t.Fatalf("list should show the label and never the token: %d %s", listed.Code, listed.Body.String())
	}

	path := "/auth/api-token/" + strconv.FormatInt(id, 10)
	decode(t, api.do(t, http.MethodDelete, path, other, nil), http.StatusNotFound)
	decode(t, api.do(t, http.MethodDelete, "/auth/api-token/1", session, nil), http.StatusNotFound)
	decode(t, api.do(t, http.MethodDelete, path, session, nil), http.StatusOK)
	decode(t, api.do(t, http.MethodGet, "/auth/me", token, nil), http.StatusUnauthorized)
	decode(t, api.do(t, http.MethodGet, "/auth/me", session, nil), http.StatusOK)
}

package mcp_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestPendingPageConfirmsOnlyFromTheSession opens a send through a client that cannot
// ask, refuses it to a delegated token, then confirms it from the person's session.
func TestPendingPageConfirmsOnlyFromTheSession(t *testing.T) {
	f := newFixture(t)
	doc := f.draft(t, "Offer")
	var out struct {
		Status     string `json:"status"`
		ConfirmURL string `json:"confirm_url"`
	}
	call(t, f.connect(t, "owner"), "plume_send_document", map[string]any{"document_id": doc.Document.ID}, &out)
	if out.Status != "pending_confirmation" || f.status(doc.Document.ID) != "draft" {
		t.Fatalf("a client that cannot ask must get a link and send nothing, got %+v", out)
	}
	actionID := out.ConfirmURL[strings.LastIndex(out.ConfirmURL, "=")+1:]
	base := strings.TrimSuffix(f.resource, "/mcp") + "/api/mcp/pending"
	checkPageLists(t, base, f.users["owner"], actionID)
	if res := pageCall(t, http.MethodPost, base+"/"+actionID+"/confirm", f.users["owner"], "bulle"); res.StatusCode != http.StatusForbidden {
		t.Fatalf("a delegated token must not confirm, got %d", res.StatusCode)
	}
	checkNotOwnSessionRefused(t, base+"/"+actionID+"/confirm", f.users["owner"])
	if res := pageCall(t, http.MethodPost, base+"/"+actionID+"/confirm", f.users["stranger"], ""); res.StatusCode == http.StatusOK {
		t.Fatal("a stranger must not confirm the owner's action")
	}
	res := pageCall(t, http.MethodPost, base+"/"+actionID+"/confirm", f.users["owner"], "")
	if res.StatusCode != http.StatusOK || f.status(doc.Document.ID) != "pending" {
		t.Fatalf("the person's session must send the draft, got %d and %s", res.StatusCode, f.status(doc.Document.ID))
	}
}

// checkNotOwnSessionRefused refuses a confirm from the person's CLI or API token session
// and from a browser cookie sent next to an Authorization header.
func checkNotOwnSessionRefused(t *testing.T, url string, userID int64) {
	t.Helper()
	for name, header := range map[string][2]string{
		"a CLI session":           {"X-Test-Session", "cli"},
		"an Authorization header": {"Authorization", "Bearer token"},
	} {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, nil)
		must(t, err)
		req.Header.Set("X-Test-User", itoa(userID))
		req.Header.Set(header[0], header[1])
		res, err := http.DefaultClient.Do(req)
		must(t, err)
		res.Body.Close()
		if res.StatusCode != http.StatusForbidden {
			t.Fatalf("%s must not confirm, got %d", name, res.StatusCode)
		}
	}
}

// checkPageLists finds the one open action on userID's page, with its document.
func checkPageLists(t *testing.T, base string, userID int64, actionID string) {
	t.Helper()
	var list struct {
		Actions []struct {
			ID       string `json:"id"`
			State    string `json:"state"`
			Document struct {
				Name    string   `json:"name"`
				Signers []string `json:"signers"`
			} `json:"document"`
		} `json:"actions"`
	}
	decode(t, pageCall(t, http.MethodGet, base+"/", userID, ""), &list)
	if len(list.Actions) != 1 || list.Actions[0].ID != actionID || list.Actions[0].State != "ready" || len(list.Actions[0].Document.Signers) != 1 {
		t.Fatalf("the page must list the send with its document, got %+v", list)
	}
}

// pageCall calls the confirm page API as userID, delegated to actor when set.
func pageCall(t *testing.T, method, url string, userID int64, actor string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, url, nil)
	must(t, err)
	req.Header.Set("X-Test-User", itoa(userID))
	if actor != "" {
		req.Header.Set("X-Test-Actor", actor)
	}
	res, err := http.DefaultClient.Do(req)
	must(t, err)
	t.Cleanup(func() { res.Body.Close() })
	return res
}

// decode reads a JSON body.
func decode(t *testing.T, res *http.Response, out any) {
	t.Helper()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	must(t, json.NewDecoder(res.Body).Decode(out))
}

// TestServe serves the endpoint on MCP_SERVE for the official conformance suite, with
// the owner's token added to every request that carries none, until interrupted.
func TestServe(t *testing.T) {
	addr := envServe()
	if addr == "" {
		t.Skip("set MCP_SERVE=127.0.0.1:<port> to serve the endpoint")
	}
	f := newFixtureOn(t, addr)
	token := f.token(t)
	f.front = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") == "" {
				r.Header.Set("Authorization", "Bearer "+token)
			}
			next.ServeHTTP(w, r)
		})
	}
	f.draft(t, "Served")
	t.Logf("serving %s", f.resource)
	waitInterrupt()
}

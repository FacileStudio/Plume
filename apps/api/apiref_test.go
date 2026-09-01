package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/modules/auth"
	"github.com/FacileStudio/porte/local"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/go-chi/chi/v5"
)

// testRouter builds the real router against nil dependencies. Route
// registration never touches the database, so the shape of the router is
// faithful even though no handler could serve a request.
//
// porte owns /auth/config, /auth/logout and the OIDC flow now, so the
// router has to carry it or this guard passes over the routes most
// likely to move. It is built over a nil database like everything else
// here: registration never reads one.
func testRouter(t *testing.T) chi.Router {
	t.Helper()
	appEnv := env.Config{}
	logger := slog.Default()

	store := portepg.New(nil)
	sessions, err := session.New(appEnv.Porte(), session.Deps{Sessions: store.Sessions(), Logger: logger})
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}
	kit, err := oidc.New(context.Background(), appEnv.Porte(), oidc.Deps{Sessions: sessions, Logger: logger})
	if err != nil {
		t.Fatalf("oidc.New: %v", err)
	}
	passwords, err := local.New(local.Config{}, local.Deps{
		Users:      auth.NewUserStore(nil),
		Identities: store.Identities(),
		Sessions:   sessions,
		Count:      func(context.Context) (int64, error) { return 0, nil },
	})
	if err != nil {
		t.Fatalf("local.New: %v", err)
	}
	return buildRouter(newServices(nil, appEnv, logger, sessions, passwords), nil, appEnv, logger, sessions, kit)
}

func TestEveryRouteIsDocumented(t *testing.T) {
	if missing := apiref.Undocumented(testRouter(t), apiReference()); len(missing) > 0 {
		t.Errorf("routes missing from the API registry: %v", missing)
	}
}

func TestRegistryIsComplete(t *testing.T) {
	if issues := apiref.Incomplete(apiReference(),
		"/auth/logout",
		"/documents",
		"/documents/{id}",
		"/documents/{id}/file",
		"/documents/{id}/send",
		"/documents/{docId}/fields/{fieldId}",
		"/documents/{docId}/certificate",
		"/documents/{docId}/audit-trail",
		"/signers/{id}",
		"/sign/{token}/opened.gif",
		"/sign/{token}/file",
		"/sign/{token}",
		"/spaces/{spaceId}",
		"/spaces/{spaceId}/leave",
		"/spaces/{spaceId}/members/{memberId}",
		"/clients/{clientId}",
		"/smtp",
		"/smtp/test",
		"/verify",
		"/webhooks/{id}",
		"/webhooks/{id}/test",
	); len(issues) > 0 {
		t.Errorf("incomplete documentation entries:\n%s", strings.Join(issues, "\n"))
	}
}

func TestReferenceIsServedAtRoot(t *testing.T) {
	page := httptest.NewRecorder()
	testRouter(t).ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if page.Code != http.StatusOK {
		t.Fatalf("GET /docs = %d, want 200", page.Code)
	}
	if !strings.Contains(page.Body.String(), `data-url="/docs/openapi.json"`) {
		t.Errorf("reference page does not point at its own spec:\n%s", page.Body.String())
	}

	spec := httptest.NewRecorder()
	testRouter(t).ServeHTTP(spec, httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil))
	if spec.Code != http.StatusOK {
		t.Fatalf("GET /docs/openapi.json = %d, want 200", spec.Code)
	}
}

func TestPublicRoutesCarryNoSecurity(t *testing.T) {
	document := apiref.OpenAPI(apiReference())
	paths := document["paths"].(map[string]any)

	login := paths["/auth/login"].(map[string]any)["post"].(map[string]any)
	if _, secured := login["security"]; secured {
		t.Error("/auth/login is public but the document demands a bearer token")
	}

	params := paths["/documents/{id}"].(map[string]any)["get"].(map[string]any)["parameters"].([]any)
	schema := params[0].(map[string]any)["schema"].(map[string]any)
	if schema["type"] != "integer" {
		t.Errorf("int path parameter schema type = %v, want integer", schema["type"])
	}
}

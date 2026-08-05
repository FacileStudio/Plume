package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/go-chi/chi/v5"
)

// testRouter builds the real router against nil dependencies. Route
// registration never touches the database, so the shape of the router is
// faithful even though no handler could serve a request.
func testRouter() chi.Router {
	appEnv := env.Config{}
	return buildRouter(newServices(nil, appEnv, slog.Default()), nil, appEnv, slog.Default())
}

func TestEveryRouteIsDocumented(t *testing.T) {
	if missing := apiref.Undocumented(testRouter(), apiReference()); len(missing) > 0 {
		t.Errorf("routes missing from the API registry: %v", missing)
	}
}

func TestReferenceIsServedAtRoot(t *testing.T) {
	page := httptest.NewRecorder()
	testRouter().ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if page.Code != http.StatusOK {
		t.Fatalf("GET /docs = %d, want 200", page.Code)
	}
	if !strings.Contains(page.Body.String(), `data-url="/docs/openapi.json"`) {
		t.Errorf("reference page does not point at its own spec:\n%s", page.Body.String())
	}

	spec := httptest.NewRecorder()
	testRouter().ServeHTTP(spec, httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil))
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

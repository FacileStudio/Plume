package docs

import (
	"net/http"

	"api/internal/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the Scalar API reference UI and the OpenAPI spec.
// Mount it on the /api router so the paths resolve to /api/docs and
// /api/docs/openapi.json.
func RegisterRoutes(router chi.Router, spec map[string]any) {
	router.Get("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(scalarHTML))
	})
	router.Get("/docs/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		httpjson.WriteJSON(w, http.StatusOK, spec)
	})
}

const scalarHTML = `<!doctype html>
<html>
<head>
  <title>Plume API</title>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <style>body { margin: 0; }</style>
</head>
<body>
  <script id="api-reference" data-url="/api/docs/openapi.json"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`

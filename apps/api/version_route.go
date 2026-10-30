package main

import (
	"net/http"

	documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"
	"github.com/FacileStudio/Plume/apps/api/internal/version"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

type versionResponse struct {
	Version string `json:"version"`
}

// registerVersionRoute answers /version and /api/version with the version of
// the running binary. Both paths, because the public edge may forward only
// /api/*, like health.Mount does for /health.
func registerVersionRoute(router chi.Router) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		httpjson.WriteJSON(w, http.StatusOK, versionResponse{Version: version.String()})
	}
	router.Get("/version", handler)
	router.Get("/api/version", handler)
}

// versionDocumentation returns the route table entry for GET /version.
func versionDocumentation() documentation.Module {
	return documentation.Module{
		Name:        "version",
		Description: "Build identification.",
		Routes: []documentation.Route{{
			Method:       "GET",
			Path:         "/version",
			Summary:      "Running build",
			Description:  "Returns the commit the running image was built from, or dev when the build carries no stamp. Also served at /version outside the /api prefix.",
			Auth:         "",
			ResponseBody: versionResponse{},
		}},
	}
}

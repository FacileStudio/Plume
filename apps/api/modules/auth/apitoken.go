package auth

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/FacileStudio/Plume/apps/api/internal/authcontext"
	documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"
	"github.com/FacileStudio/Plume/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

const maxTokenLabelLength = 100

// APITokenRequest is the body of POST /auth/api-token.
type APITokenRequest struct {
	Label string `json:"label"`
}

// APITokenResponse carries the new token. The plaintext is shown once.
type APITokenResponse struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
	Token string `json:"token"`
}

// registerAPIToken mounts POST /auth/api-token behind RequireAuth: it mints a
// named bearer token with no expiry, for another app to call Plume with.
func registerAPIToken(router chi.Router, service *Service) {
	router.Group(func(router chi.Router) {
		router.Use(middleware.RequireAuth(service))

		router.Post("/api-token", func(w http.ResponseWriter, request *http.Request) {
			var req APITokenRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			label := strings.TrimSpace(req.Label)
			if label == "" || len(label) > maxTokenLabelLength {
				httpjson.WriteError(w, errors.Invalid("label is required and at most 100 characters"))
				return
			}
			identity, _ := authcontext.IdentityFromContext(request.Context())
			userID, _ := strconv.ParseInt(identity.UserID, 10, 64)
			token, session, err := service.Issue(request.Context(), userID, label)
			if err != nil {
				httpjson.WriteError(w, errors.Internal("failed to issue token", err))
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, APITokenResponse{ID: session.ID, Label: label, Token: token})
		})
	})
}

func apiTokenRoutes() []documentation.Route {
	return []documentation.Route{{
		Method:       "POST",
		Path:         "/auth/api-token",
		Summary:      "Create an API token",
		Description:  "Mints a named bearer token with no expiry for another app, such as Ardoise, to call Plume as this user. The token is returned once and cannot be read again.",
		Auth:         "bearer",
		RequestBody:  APITokenRequest{},
		ResponseBody: APITokenResponse{},
		Errors:       []documentation.Error{{Status: 400, Code: "invalid_argument", Description: "Missing or too long label."}},
	}}
}

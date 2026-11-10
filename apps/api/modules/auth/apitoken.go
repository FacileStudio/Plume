package auth

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"
	"github.com/FacileStudio/porte/kit"
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

// APITokenSummary is one row of GET /auth/api-tokens. The token itself is
// never listed: only its hash is stored.
type APITokenSummary struct {
	ID         int64     `json:"id"`
	Label      string    `json:"label"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

// APITokenRevoked is the body of DELETE /auth/api-token/{id}.
type APITokenRevoked struct {
	Deleted bool `json:"deleted"`
}

// registerAPIToken mounts the token routes behind RequireAuth: POST mints a
// named bearer token with no expiry for another app to call Plume with, GET
// lists the caller's tokens and DELETE revokes one by id.
func registerAPIToken(router chi.Router, service *Service) {
	router.Group(func(router chi.Router) {
		router.Use(service.RequireAuth)

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
			identity, _ := kit.FromContext(request.Context())
			userID := identity.UserID
			token, session, err := service.Issue(request.Context(), userID, label)
			if err != nil {
				httpjson.WriteError(w, errors.Internal("failed to issue token", err))
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, APITokenResponse{ID: session.ID, Label: label, Token: token})
		})

		router.Get("/api-tokens", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := kit.FromContext(request.Context())
			userID := identity.UserID
			sessions, err := service.Sessions().List(request.Context(), userID)
			if err != nil {
				httpjson.WriteError(w, errors.Internal("failed to list tokens", err))
				return
			}
			tokens := make([]APITokenSummary, 0, len(sessions))
			for _, session := range sessions {
				if session.Label == "" {
					continue
				}
				tokens = append(tokens, APITokenSummary{ID: session.ID, Label: session.Label, CreatedAt: session.CreatedAt, LastUsedAt: session.LastUsedAt})
			}
			httpjson.WriteJSON(w, http.StatusOK, tokens)
		})

		router.Delete("/api-token/{id}", func(w http.ResponseWriter, request *http.Request) {
			id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid token id"))
				return
			}
			identity, _ := kit.FromContext(request.Context())
			userID := identity.UserID
			if err := service.revokeAPIToken(request.Context(), userID, id); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, APITokenRevoked{Deleted: true})
		})
	})
}

func apiTokenRoutes() []documentation.Route {
	return []documentation.Route{{
		Method:       "GET",
		Path:         "/auth/api-tokens",
		Summary:      "List API tokens",
		Description:  "Lists the caller's named API tokens with their label, creation date and last use. The token values are never returned.",
		Auth:         "bearer",
		ResponseBody: []APITokenSummary{},
	}, {
		Method:       "DELETE",
		Path:         "/auth/api-token/{id}",
		Summary:      "Revoke an API token",
		Description:  "Ends one of the caller's named API tokens at once. A browser login cannot be revoked here.",
		Auth:         "bearer",
		ResponseBody: APITokenRevoked{},
		Errors:       []documentation.Error{{Status: 404, Code: "not_found", Description: "No API token with this id belongs to the caller."}},
	}, {
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

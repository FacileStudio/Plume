package webhooks

import (
	"net/http"
	"strconv"

	"api/internal/authcontext"
	"api/internal/errors"
	"api/internal/httpjson"
	"api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/webhooks", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Post("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)

			var req CreateWebhookRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.create(request.Context(), ownerID, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)

			resp, err := service.list(request.Context(), ownerID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Get("/{id}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)
			webhookID, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid webhook ID"))
				return
			}

			resp, err := service.get(request.Context(), ownerID, webhookID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Put("/{id}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)
			webhookID, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid webhook ID"))
				return
			}

			var req UpdateWebhookRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.update(request.Context(), ownerID, webhookID, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)
			webhookID, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid webhook ID"))
				return
			}

			err = service.delete(request.Context(), ownerID, webhookID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}

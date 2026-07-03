package clients

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
	router.Route("/clients", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Post("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)

			var req CreateClientRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.Create(request.Context(), ownerID, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)

			resp, err := service.List(request.Context(), ownerID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Route("/{clientId}", func(router chi.Router) {
			router.Get("/", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)
				clientID, err := strconv.ParseInt(chi.URLParam(request, "clientId"), 10, 64)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid client ID"))
					return
				}

				resp, err := service.Get(request.Context(), ownerID, clientID)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Put("/", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)
				clientID, err := strconv.ParseInt(chi.URLParam(request, "clientId"), 10, 64)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid client ID"))
					return
				}

				var req UpdateClientRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}

				resp, err := service.Update(request.Context(), ownerID, clientID, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Delete("/", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)
				clientID, err := strconv.ParseInt(chi.URLParam(request, "clientId"), 10, 64)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid client ID"))
					return
				}

				if err := service.Delete(request.Context(), ownerID, clientID); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			})
		})
	})
}

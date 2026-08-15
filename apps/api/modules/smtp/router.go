package smtp

import (
	stderrors "errors"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Plume/apps/api/internal/authcontext"
	"github.com/FacileStudio/Plume/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the SMTP configuration, test and delete endpoints
// behind auth.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/smtp", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)

			resp, err := service.getConfig(request.Context(), ownerID)
			if err != nil {
				var appErr *errors.Error
				if stderrors.As(err, &appErr) && appErr.Code == "not_found" {
					httpjson.WriteJSON(w, http.StatusOK, nil)
					return
				}
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Put("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)

			var req SaveConfigRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.saveConfig(request.Context(), ownerID, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Delete("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)

			err := service.deleteConfig(request.Context(), ownerID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		router.Post("/test", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)

			var req TestRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			err := service.testConfig(request.Context(), ownerID, req.To)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "sent"})
		})
	})
}

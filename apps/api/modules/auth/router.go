package auth

import (
	"net/http"

	"github.com/FacileStudio/Plume/apps/api/internal/authcontext"
	"github.com/FacileStudio/Plume/apps/api/internal/middleware"
	"github.com/FacileStudio/porte/oidc"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// registerMe mounts GET and PUT /me behind RequireAuth.
func registerMe(router chi.Router, service *Service) {
	router.Group(func(router chi.Router) {
		router.Use(middleware.RequireAuth(service))

		router.Get("/me", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.controller.getMe(request.Context(), identity.UserID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Put("/me", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var req UpdateProfileRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.updateMe(request.Context(), identity.UserID, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
}

// RegisterRoutes mounts the routes porte owns (/auth/config, /auth/logout,
// /auth/sync-profile and the OIDC flow) and Plume's profile and API token
// routes under /auth.
//
// There is no register, login or password route: porte mounts none while an
// issuer is configured, so sign-in goes through Registre only.
func RegisterRoutes(router chi.Router, service *Service, kit *oidc.Kit) {
	service.sessions.Mount(router)
	kit.Mount(router)
	router.Route("/auth", func(router chi.Router) {
		registerMe(router, service)
		registerAPIToken(router, service)
	})
}

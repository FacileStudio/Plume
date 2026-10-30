package auth

import (
	"net/http"

	"github.com/FacileStudio/Plume/apps/api/internal/authcontext"
	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// registerCredentials mounts POST /register and POST /login behind the shared
// rate limiter, or nothing when SSO-only mode is enabled. The limiter is a
// single instance shared by both routes so they drain one budget together.
func registerCredentials(router chi.Router, service *Service, appEnv env.Config) {
	if !appEnv.SSOOnly {
		authLimiter := middleware.NewRateLimiter(10, 5).Handler()

		router.Group(func(router chi.Router) {
			router.Use(authLimiter)
			registerRegister(router, service)
			registerLogin(router, service)
		})
	}
}

// registerRegister mounts the local account-creation route.
func registerRegister(router chi.Router, service *Service) {
	router.Post("/register", func(w http.ResponseWriter, request *http.Request) {
		var req RegisterRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		resp, err := service.controller.register(w, request, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, resp)
	})
}

// registerLogin mounts the local credential-authentication route.
func registerLogin(router chi.Router, service *Service) {
	router.Post("/login", func(w http.ResponseWriter, request *http.Request) {
		var req LoginRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		resp, err := service.controller.login(w, request, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

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

// registerPassword mounts PUT /password behind RequireAuth.
func registerPassword(router chi.Router, service *Service) {
	router.Group(func(router chi.Router) {
		router.Use(middleware.RequireAuth(service))

		router.Put("/password", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var req ChangePasswordRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.changePassword(w, request, identity.UserID, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
}

// RegisterRoutes mounts /auth registration, login, profile, and password
// endpoints on router. Registration and login are skipped when SSO-only mode
// is enabled and rate-limited otherwise; profile and password routes require
// auth.
func RegisterRoutes(router chi.Router, service *Service, appEnv env.Config) {
	router.Route("/auth", func(router chi.Router) {
		registerCredentials(router, service, appEnv)
		registerMe(router, service)
		registerPassword(router, service)
		registerAPIToken(router, service)
	})
}

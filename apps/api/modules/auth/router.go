package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"api/internal/authcontext"
	"api/internal/env"
	"api/internal/httpjson"
	"api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, appEnv env.Config) {
	oidcEnabled := appEnv.OIDC != nil

	router.Route("/auth", func(router chi.Router) {
		router.Get("/config", func(w http.ResponseWriter, r *http.Request) {
			cfg := map[string]any{
				"sso_only":     appEnv.SSOOnly,
				"oidc_enabled": oidcEnabled,
			}
			if oidcEnabled {
				cfg["oidc_redirect_url"] = appEnv.OIDC.RedirectURL
				cfg["oidc_issuer"] = appEnv.OIDC.Issuer
			}
			httpjson.WriteJSON(w, http.StatusOK, cfg)
		})

		if !appEnv.SSOOnly {
			authLimiter := middleware.NewRateLimiter(10, 5).Handler()

			router.Group(func(router chi.Router) {
				router.Use(authLimiter)

				router.Post("/register", func(w http.ResponseWriter, request *http.Request) {
					var req RegisterRequest
					if err := httpjson.DecodeJSON(w, request, &req); err != nil {
						httpjson.WriteError(w, err)
						return
					}
					resp, err := service.controller.register(request.Context(), &req)
					if err != nil {
						httpjson.WriteError(w, err)
						return
					}
					httpjson.WriteJSON(w, http.StatusCreated, resp)
				})

				router.Post("/login", func(w http.ResponseWriter, request *http.Request) {
					var req LoginRequest
					if err := httpjson.DecodeJSON(w, request, &req); err != nil {
						httpjson.WriteError(w, err)
						return
					}
					resp, err := service.controller.login(request.Context(), &req)
					if err != nil {
						httpjson.WriteError(w, err)
						return
					}
					httpjson.WriteJSON(w, http.StatusOK, resp)
				})
			})
		}

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

			router.Put("/password", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				var req ChangePasswordRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				if err := service.controller.changePassword(request.Context(), identity.UserID, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			})
		})

		if oidcEnabled {
			oidc, err := newOIDCHandler(context.Background(), appEnv.OIDC, service)
			if err != nil {
				slog.Error("failed to initialize OIDC provider", slog.Any("error", err))
			} else {
				go func() {
					ticker := time.NewTicker(5 * time.Minute)
					defer ticker.Stop()
					for range ticker.C {
						now := time.Now()
						oidc.codes.Range(func(key, value any) bool {
							if pending, ok := value.(pendingCode); ok && now.After(pending.ExpiresAt) {
								oidc.codes.Delete(key)
							}
							return true
						})
					}
				}()

				router.Get("/oidc", oidc.login)
				router.Get("/oidc/callback", oidc.callback)
				router.Post("/oidc/exchange", oidc.exchange)
			}
		}
	})
}

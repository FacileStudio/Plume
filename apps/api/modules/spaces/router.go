package spaces

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
	router.Route("/spaces", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Post("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)

			var req CreateSpaceRequest
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
			userID, _ := strconv.ParseInt(identity.UserID, 10, 64)

			resp, err := service.List(request.Context(), userID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Route("/{spaceId}", func(router chi.Router) {
			router.Get("/", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				userID, _ := strconv.ParseInt(identity.UserID, 10, 64)
				spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid space ID"))
					return
				}

				resp, err := service.Get(request.Context(), userID, spaceID)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Put("/", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				userID, _ := strconv.ParseInt(identity.UserID, 10, 64)
				spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid space ID"))
					return
				}

				var req UpdateSpaceRequest
				if err := httpjson.DecodeJSON(w, request, &req); err != nil {
					httpjson.WriteError(w, err)
					return
				}

				resp, err := service.Update(request.Context(), userID, spaceID, &req)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Delete("/", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				userID, _ := strconv.ParseInt(identity.UserID, 10, 64)
				spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid space ID"))
					return
				}

				if err := service.Delete(request.Context(), userID, spaceID); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			})

			router.Post("/leave", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				userID, _ := strconv.ParseInt(identity.UserID, 10, 64)
				spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid space ID"))
					return
				}

				if err := service.Leave(request.Context(), userID, spaceID); err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			})

			router.Route("/members", func(router chi.Router) {
				router.Get("/", func(w http.ResponseWriter, request *http.Request) {
					identity, _ := authcontext.IdentityFromContext(request.Context())
					userID, _ := strconv.ParseInt(identity.UserID, 10, 64)
					spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
					if err != nil {
						httpjson.WriteError(w, errors.Invalid("invalid space ID"))
						return
					}

					resp, err := service.ListMembers(request.Context(), userID, spaceID)
					if err != nil {
						httpjson.WriteError(w, err)
						return
					}
					httpjson.WriteJSON(w, http.StatusOK, resp)
				})

				router.Post("/", func(w http.ResponseWriter, request *http.Request) {
					identity, _ := authcontext.IdentityFromContext(request.Context())
					userID, _ := strconv.ParseInt(identity.UserID, 10, 64)
					spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
					if err != nil {
						httpjson.WriteError(w, errors.Invalid("invalid space ID"))
						return
					}

					var req AddMemberRequest
					if err := httpjson.DecodeJSON(w, request, &req); err != nil {
						httpjson.WriteError(w, err)
						return
					}

					resp, err := service.AddMember(request.Context(), userID, spaceID, &req)
					if err != nil {
						httpjson.WriteError(w, err)
						return
					}
					httpjson.WriteJSON(w, http.StatusCreated, resp)
				})

				router.Put("/{memberId}", func(w http.ResponseWriter, request *http.Request) {
					identity, _ := authcontext.IdentityFromContext(request.Context())
					userID, _ := strconv.ParseInt(identity.UserID, 10, 64)
					spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
					if err != nil {
						httpjson.WriteError(w, errors.Invalid("invalid space ID"))
						return
					}
					memberID, err := strconv.ParseInt(chi.URLParam(request, "memberId"), 10, 64)
					if err != nil {
						httpjson.WriteError(w, errors.Invalid("invalid member ID"))
						return
					}

					var req UpdateMemberRoleRequest
					if err := httpjson.DecodeJSON(w, request, &req); err != nil {
						httpjson.WriteError(w, err)
						return
					}

					resp, err := service.UpdateMemberRole(request.Context(), userID, spaceID, memberID, &req)
					if err != nil {
						httpjson.WriteError(w, err)
						return
					}
					httpjson.WriteJSON(w, http.StatusOK, resp)
				})

				router.Delete("/{memberId}", func(w http.ResponseWriter, request *http.Request) {
					identity, _ := authcontext.IdentityFromContext(request.Context())
					userID, _ := strconv.ParseInt(identity.UserID, 10, 64)
					spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
					if err != nil {
						httpjson.WriteError(w, errors.Invalid("invalid space ID"))
						return
					}
					memberID, err := strconv.ParseInt(chi.URLParam(request, "memberId"), 10, 64)
					if err != nil {
						httpjson.WriteError(w, errors.Invalid("invalid member ID"))
						return
					}

					if err := service.RemoveMember(request.Context(), userID, spaceID, memberID); err != nil {
						httpjson.WriteError(w, err)
						return
					}
					w.WriteHeader(http.StatusNoContent)
				})
			})
		})
	})
}

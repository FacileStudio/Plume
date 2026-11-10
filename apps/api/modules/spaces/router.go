package spaces

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// ownerIDFrom extracts the authenticated user's numeric ID from the request.
// The bool reports whether an identity was present; routes guarded by
// RequireAuth can ignore it.
func ownerIDFrom(r *http.Request) (int64, bool) {
	identity, ok := kit.FromContext(r.Context())
	return identity.UserID, ok
}

// RegisterRoutes mounts the spaces CRUD and membership endpoints behind auth.
func RegisterRoutes(router chi.Router, service *Service, requireAuth func(http.Handler) http.Handler) {
	router.Route("/spaces", func(router chi.Router) {
		router.Use(requireAuth)

		RegisterSpaceCollectionRoutes(router, service)
		RegisterSpaceGetRoute(router, service)
		RegisterSpaceLeaveRoute(router, service)
		RegisterSpaceUpdateRoute(router, service)
		RegisterSpaceDeleteRoute(router, service)
		RegisterMemberListRoute(router, service)
		RegisterMemberAddRoute(router, service)
		RegisterMemberUpdateRoute(router, service)
		RegisterMemberRemoveRoute(router, service)
	})
}

// RegisterSpaceCollectionRoutes mounts POST /spaces and GET /spaces.
func RegisterSpaceCollectionRoutes(router chi.Router, service *Service) {
	router.Post("/", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)

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
		ownerID, _ := ownerIDFrom(request)

		resp, err := service.List(request.Context(), ownerID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterSpaceGetRoute mounts GET /spaces/{spaceId}.
func RegisterSpaceGetRoute(router chi.Router, service *Service) {
	router.Get("/{spaceId}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid space ID"))
			return
		}

		resp, err := service.Get(request.Context(), ownerID, spaceID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterSpaceLeaveRoute mounts POST /spaces/{spaceId}/leave.
func RegisterSpaceLeaveRoute(router chi.Router, service *Service) {
	router.Post("/{spaceId}/leave", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid space ID"))
			return
		}

		if err := service.Leave(request.Context(), ownerID, spaceID); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}

// RegisterSpaceUpdateRoute mounts PUT /spaces/{spaceId}.
func RegisterSpaceUpdateRoute(router chi.Router, service *Service) {
	router.Put("/{spaceId}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
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

		resp, err := service.Update(request.Context(), ownerID, spaceID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterSpaceDeleteRoute mounts DELETE /spaces/{spaceId}.
func RegisterSpaceDeleteRoute(router chi.Router, service *Service) {
	router.Delete("/{spaceId}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid space ID"))
			return
		}

		if err := service.Delete(request.Context(), ownerID, spaceID); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// RegisterSyncRoute mounts Registre's nudge at POST /registre/sync, outside
// RequireAuth: Registre posts it anonymously, and all it does is make the
// next authenticated pull happen sooner. A nil registre, a Plume without
// OIDC, mounts nothing.
func RegisterSyncRoute(router chi.Router, registre *Registre) {
	if registre == nil {
		return
	}
	router.Method(http.MethodPost, strings.TrimPrefix(directory.SyncPath, "/api"), registre.syncer.Handler())
}

// RegisterDeletionsRoute mounts Registre's signed user.deleted push at
// POST /registre/deletions, outside RequireAuth: the handler verifies the
// notice's signature itself. A nil deletions, a Plume without OIDC, mounts
// nothing.
func RegisterDeletionsRoute(router chi.Router, deletions *Deletions) {
	if deletions == nil {
		return
	}
	router.Method(http.MethodPost, strings.TrimPrefix(directory.DeletionsPath, "/api"), deletions.Handler())
}

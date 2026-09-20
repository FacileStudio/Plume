package fields

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Plume/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// ownerIDFrom extracts the authenticated owner's numeric ID from the request.
func ownerIDFrom(r *http.Request) (int64, bool) {
	identity, _ := authcontext.IdentityFromContext(r.Context())
	ownerID, err := strconv.ParseInt(identity.UserID, 10, 64)
	return ownerID, err == nil
}

// DocumentRoutes returns a router mount function adding field CRUD routes
// under a document's /{docId}/fields path.
func DocumentRoutes(service *Service) func(chi.Router) {
	return func(router chi.Router) {
		RegisterFieldListRoute(router, service)
		RegisterFieldCreateRoute(router, service)
		RegisterFieldUpdateRoute(router, service)
		RegisterFieldDeleteRoute(router, service)
	}
}

// RegisterFieldListRoute mounts GET /{docId}/fields.
func RegisterFieldListRoute(router chi.Router, service *Service) {
	router.Get("/{docId}/fields", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)

		resp, err := service.List(request.Context(), strconv.FormatInt(ownerID, 10), chi.URLParam(request, "docId"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterFieldCreateRoute mounts POST /{docId}/fields.
func RegisterFieldCreateRoute(router chi.Router, service *Service) {
	router.Post("/{docId}/fields", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)

		var req CreateFieldRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}

		resp, err := service.Create(request.Context(), strconv.FormatInt(ownerID, 10), chi.URLParam(request, "docId"), &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, resp)
	})
}

// RegisterFieldUpdateRoute mounts PUT /{docId}/fields/{fieldId}.
func RegisterFieldUpdateRoute(router chi.Router, service *Service) {
	router.Put("/{docId}/fields/{fieldId}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)

		var req UpdateFieldRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}

		resp, err := service.Update(request.Context(), strconv.FormatInt(ownerID, 10), chi.URLParam(request, "docId"), chi.URLParam(request, "fieldId"), &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterFieldDeleteRoute mounts DELETE /{docId}/fields/{fieldId}.
func RegisterFieldDeleteRoute(router chi.Router, service *Service) {
	router.Delete("/{docId}/fields/{fieldId}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)

		err := service.Delete(request.Context(), strconv.FormatInt(ownerID, 10), chi.URLParam(request, "docId"), chi.URLParam(request, "fieldId"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

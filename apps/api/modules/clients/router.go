package clients

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// ownerIDFrom extracts the authenticated owner's numeric ID from the request.
func ownerIDFrom(r *http.Request) (int64, bool) {
	identity, ok := kit.FromContext(r.Context())
	return identity.UserID, ok
}

// RegisterRoutes mounts the /clients CRUD routes on router, all requiring
// auth and scoped to the authenticated owner.
func RegisterRoutes(router chi.Router, service *Service, requireAuth func(http.Handler) http.Handler) {
	router.Route("/clients", func(router chi.Router) {
		router.Use(requireAuth)

		RegisterClientCollectionRoutes(router, service)
		RegisterClientItemRoutes(router, service)
	})
}

// RegisterClientCollectionRoutes mounts the /clients collection routes.
func RegisterClientCollectionRoutes(router chi.Router, service *Service) {
	router.Post("/", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)

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
		ownerID, _ := ownerIDFrom(request)

		resp, err := service.List(request.Context(), ownerID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterClientItemRoutes mounts the /clients/{clientId} item routes.
func RegisterClientItemRoutes(router chi.Router, service *Service) {
	router.Route("/{clientId}", func(router chi.Router) {
		RegisterClientGetRoute(router, service)
		RegisterClientUpdateRoute(router, service)
		RegisterClientDeleteRoute(router, service)
	})
}

// RegisterClientGetRoute mounts GET /clients/{clientId}.
func RegisterClientGetRoute(router chi.Router, service *Service) {
	router.Get("/", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
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
}

// RegisterClientUpdateRoute mounts PUT /clients/{clientId}.
func RegisterClientUpdateRoute(router chi.Router, service *Service) {
	router.Put("/", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
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
}

// RegisterClientDeleteRoute mounts DELETE /clients/{clientId}.
func RegisterClientDeleteRoute(router chi.Router, service *Service) {
	router.Delete("/", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
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
}

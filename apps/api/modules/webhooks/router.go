package webhooks

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

// RegisterRoutes mounts the webhook CRUD and test endpoints behind auth.
func RegisterRoutes(router chi.Router, service *Service, requireAuth func(http.Handler) http.Handler) {
	router.Route("/webhooks", func(router chi.Router) {
		router.Use(requireAuth)

		RegisterWebhookCollectionRoutes(router, service)
		RegisterWebhookGetRoute(router, service)
		RegisterWebhookUpdateRoute(router, service)
		RegisterWebhookDeleteRoute(router, service)
		RegisterWebhookTestRoute(router, service)
	})
}

// RegisterWebhookCollectionRoutes mounts the POST and GET /webhooks routes.
func RegisterWebhookCollectionRoutes(router chi.Router, service *Service) {
	router.Post("/", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)

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
		ownerID, _ := ownerIDFrom(request)

		resp, err := service.list(request.Context(), ownerID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterWebhookGetRoute mounts GET /webhooks/{id}.
func RegisterWebhookGetRoute(router chi.Router, service *Service) {
	router.Get("/{id}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
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
}

// RegisterWebhookUpdateRoute mounts PUT /webhooks/{id}.
func RegisterWebhookUpdateRoute(router chi.Router, service *Service) {
	router.Put("/{id}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
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
}

// RegisterWebhookDeleteRoute mounts DELETE /webhooks/{id}.
func RegisterWebhookDeleteRoute(router chi.Router, service *Service) {
	router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
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
}

// RegisterWebhookTestRoute mounts POST /webhooks/{id}/test.
func RegisterWebhookTestRoute(router chi.Router, service *Service) {
	router.Post("/{id}/test", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		webhookID, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid webhook ID"))
			return
		}

		if err := service.Test(request.Context(), ownerID, webhookID); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "delivered"})
	})
}

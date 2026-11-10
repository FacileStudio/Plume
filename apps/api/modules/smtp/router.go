package smtp

import (
	stderrors "errors"
	"net/http"

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

// RegisterRoutes mounts the SMTP configuration, test and delete endpoints
// behind auth.
func RegisterRoutes(router chi.Router, service *Service, requireAuth func(http.Handler) http.Handler) {
	router.Route("/smtp", func(router chi.Router) {
		router.Use(requireAuth)

		RegisterSmtpGetRoute(router, service)
		RegisterSmtpSaveRoute(router, service)
		RegisterSmtpDeleteRoute(router, service)
		RegisterSmtpTestRoute(router, service)
	})
}

// RegisterSmtpGetRoute mounts GET /smtp.
func RegisterSmtpGetRoute(router chi.Router, service *Service) {
	router.Get("/", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)

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
}

// RegisterSmtpSaveRoute mounts PUT /smtp.
func RegisterSmtpSaveRoute(router chi.Router, service *Service) {
	router.Put("/", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)

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
}

// RegisterSmtpDeleteRoute mounts DELETE /smtp.
func RegisterSmtpDeleteRoute(router chi.Router, service *Service) {
	router.Delete("/", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)

		err := service.deleteConfig(request.Context(), ownerID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// RegisterSmtpTestRoute mounts POST /smtp/test.
func RegisterSmtpTestRoute(router chi.Router, service *Service) {
	router.Post("/test", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)

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
}

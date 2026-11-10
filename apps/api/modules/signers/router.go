package signers

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// trackingPixelGIF is a 1x1 fully transparent GIF used as an email open beacon.
var trackingPixelGIF = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00,
	0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x21, 0xF9, 0x04, 0x01, 0x00, 0x00, 0x00,
	0x00, 0x2C, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02,
	0x44, 0x01, 0x00, 0x3B,
}

// ownerIDFrom extracts the authenticated user's ID from the request context,
// parsed as an integer. The bool reports whether an identity was present;
// routes guarded by RequireAuth can ignore it.
func ownerIDFrom(r *http.Request) (int64, bool) {
	identity, ok := kit.FromContext(r.Context())
	return identity.UserID, ok
}

// clientIP returns the request's client IP, honouring X-Real-IP when a trusted
// proxy has forwarded the connection.
func clientIP(request *http.Request) string {
	ip := request.RemoteAddr
	if forwarded := request.Header.Get("X-Real-IP"); forwarded != "" {
		ip = forwarded
	}
	return ip
}

// DocumentRoutes returns the router that mounts the signer-management
// endpoints (list, add) under a document-scoped subrouter.
func DocumentRoutes(service *Service) func(chi.Router) {
	return func(router chi.Router) {
		router.Get("/{docId}/signers", func(w http.ResponseWriter, request *http.Request) {
			ownerID, _ := ownerIDFrom(request)
			resp, err := service.ListSigners(request.Context(), strconv.FormatInt(ownerID, 10), chi.URLParam(request, "docId"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/{docId}/signers", func(w http.ResponseWriter, request *http.Request) {
			ownerID, _ := ownerIDFrom(request)
			var req AddSignerRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.AddSigner(request.Context(), strconv.FormatInt(ownerID, 10), chi.URLParam(request, "docId"), &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})
	}
}

// RegisterSignerRoutes mounts the authenticated /signers resource: removing a
// signer, plus any nested routes such as reminders' send-remind action.
func RegisterSignerRoutes(router chi.Router, service *Service, requireAuth func(http.Handler) http.Handler, nested ...func(chi.Router)) {
	router.Route("/signers", func(router chi.Router) {
		router.Use(requireAuth)

		router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
			ownerID, _ := ownerIDFrom(request)
			err := service.RemoveSigner(request.Context(), strconv.FormatInt(ownerID, 10), chi.URLParam(request, "id"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		for _, fn := range nested {
			fn(router)
		}
	})
}

// RegisterSigningStateRoutes mounts the public token-scoped endpoints that
// read signing state: the signing view and the progress roster.
func RegisterSigningStateRoutes(router chi.Router, service *Service) {
	router.Get("/sign/{token}", func(w http.ResponseWriter, request *http.Request) {
		resp, err := service.GetSigningView(request.Context(), chi.URLParam(request, "token"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Get("/sign/{token}/status", func(w http.ResponseWriter, request *http.Request) {
		resp, err := service.GetSigningStatus(request.Context(), chi.URLParam(request, "token"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterSigningDeliveryRoutes mounts the public token-scoped endpoints that
// serve raw resources: the email-open tracking pixel and the PDF download.
func RegisterSigningDeliveryRoutes(router chi.Router, service *Service) {
	router.Get("/sign/{token}/opened.gif", func(w http.ResponseWriter, request *http.Request) {
		service.MarkEmailOpened(request.Context(), chi.URLParam(request, "token"))
		w.Header().Set("Content-Type", "image/gif")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Length", strconv.Itoa(len(trackingPixelGIF)))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(trackingPixelGIF); err != nil {
			return
		}
	})

	router.Get("/sign/{token}/file", func(w http.ResponseWriter, request *http.Request) {
		filePath, err := service.GetSigningFilePath(request.Context(), chi.URLParam(request, "token"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeFile(w, request, filePath)
	})
}

// RegisterSigningDecisionRoutes mounts the public token-scoped endpoints that
// move a signature forward: submitting the field values and declining.
func RegisterSigningDecisionRoutes(router chi.Router, service *Service) {
	router.Post("/sign/{token}", func(w http.ResponseWriter, request *http.Request) {
		var req SubmitSignatureRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		err := service.SubmitSignature(request.Context(), chi.URLParam(request, "token"), &req, clientIP(request), request.UserAgent())
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "signed"})
	})

	router.Post("/sign/{token}/decline", func(w http.ResponseWriter, request *http.Request) {
		err := service.DeclineSignature(request.Context(), chi.URLParam(request, "token"), clientIP(request), request.UserAgent())
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "declined"})
	})
}

// RegisterRoutes mounts the signer CRUD and the public, token-scoped signing
// endpoints, applying auth to everything except the token sign routes.
func RegisterRoutes(router chi.Router, service *Service, requireAuth func(http.Handler) http.Handler, nested ...func(chi.Router)) {
	RegisterSignerRoutes(router, service, requireAuth, nested...)
	RegisterSigningStateRoutes(router, service)
	RegisterSigningDeliveryRoutes(router, service)
	RegisterSigningDecisionRoutes(router, service)
}

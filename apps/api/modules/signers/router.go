package signers

import (
	"net/http"
	"strconv"

	"api/internal/authcontext"
	"api/internal/httpjson"
	"api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

// trackingPixelGIF is a 1x1 fully transparent GIF used as an email open beacon.
var trackingPixelGIF = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00,
	0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x21, 0xF9, 0x04, 0x01, 0x00, 0x00, 0x00,
	0x00, 0x2C, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02,
	0x44, 0x01, 0x00, 0x3B,
}

func DocumentRoutes(service *Service) func(chi.Router) {
	return func(router chi.Router) {
		router.Get("/{docId}/signers", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.ListSigners(request.Context(), identity.UserID, chi.URLParam(request, "docId"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/{docId}/signers", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var req AddSignerRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.AddSigner(request.Context(), identity.UserID, chi.URLParam(request, "docId"), &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})
	}
}

func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator, nested ...func(chi.Router)) {
	router.Route("/signers", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			err := service.RemoveSigner(request.Context(), identity.UserID, chi.URLParam(request, "id"))
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

	router.Get("/sign/{token}", func(w http.ResponseWriter, request *http.Request) {
		resp, err := service.GetSigningView(request.Context(), chi.URLParam(request, "token"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Get("/sign/{token}/opened.gif", func(w http.ResponseWriter, request *http.Request) {
		service.MarkEmailOpened(request.Context(), chi.URLParam(request, "token"))
		w.Header().Set("Content-Type", "image/gif")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Length", strconv.Itoa(len(trackingPixelGIF)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(trackingPixelGIF)
	})

	router.Post("/sign/{token}", func(w http.ResponseWriter, request *http.Request) {
		var req SubmitSignatureRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		ip := request.RemoteAddr
		if forwarded := request.Header.Get("X-Real-IP"); forwarded != "" {
			ip = forwarded
		}
		err := service.SubmitSignature(request.Context(), chi.URLParam(request, "token"), &req, ip, request.UserAgent())
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "signed"})
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

	router.Post("/sign/{token}/decline", func(w http.ResponseWriter, request *http.Request) {
		ip := request.RemoteAddr
		if forwarded := request.Header.Get("X-Real-IP"); forwarded != "" {
			ip = forwarded
		}
		err := service.DeclineSignature(request.Context(), chi.URLParam(request, "token"), ip, request.UserAgent())
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "declined"})
	})
}

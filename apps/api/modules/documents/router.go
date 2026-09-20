package documents

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Plume/apps/api/internal/authcontext"
	"github.com/FacileStudio/Plume/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the /documents routes on router: upload, list,
// stats, get, file download, update, delete, and send. All require auth.
// Additional nested routers can be mounted under /documents via nested.
func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator, nested ...func(chi.Router)) {
	router.Route("/documents", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		for _, fn := range nested {
			fn(router)
		}

		RegisterUploadRoute(router, service)
		RegisterListRoutes(router, service)
		RegisterGetRoute(router, service)
		RegisterFileRoute(router, service)
		RegisterUpdateRoute(router, service)
		RegisterDeleteRoute(router, service)
		RegisterSendRoute(router, service)
	})
}

// ownerIDFrom extracts the authenticated user's numeric ID from the request.
// The bool reports whether the identity was present and parsed.
func ownerIDFrom(r *http.Request) (int64, bool) {
	identity, _ := authcontext.IdentityFromContext(r.Context())
	ownerID, err := strconv.ParseInt(identity.UserID, 10, 64)
	return ownerID, err == nil
}

// RegisterListRoutes mounts the collection read routes: GET /documents/stats
// and GET /documents.
func RegisterListRoutes(router chi.Router, service *Service) {
	router.Get("/stats", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		spaceID := request.URL.Query().Get("space_id")
		resp, err := service.Stats(request.Context(), strconv.FormatInt(ownerID, 10), spaceID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})

	router.Get("/", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		status := request.URL.Query().Get("status")
		spaceID := request.URL.Query().Get("space_id")
		clientID := request.URL.Query().Get("client_id")
		resp, err := service.List(request.Context(), strconv.FormatInt(ownerID, 10), status, spaceID, clientID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterGetRoute mounts GET /documents/{id}.
func RegisterGetRoute(router chi.Router, service *Service) {
	router.Get("/{id}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		resp, err := service.Get(request.Context(), strconv.FormatInt(ownerID, 10), chi.URLParam(request, "id"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterFileRoute mounts GET /documents/{id}/file, streaming the PDF.
func RegisterFileRoute(router chi.Router, service *Service) {
	router.Get("/{id}/file", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		filePath, err := service.GetFilePath(request.Context(), strconv.FormatInt(ownerID, 10), chi.URLParam(request, "id"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeFile(w, request, filePath)
	})
}

// RegisterUpdateRoute mounts PUT /documents/{id}.
func RegisterUpdateRoute(router chi.Router, service *Service) {
	router.Put("/{id}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		var req UpdateRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		resp, err := service.Update(request.Context(), strconv.FormatInt(ownerID, 10), chi.URLParam(request, "id"), &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterDeleteRoute mounts DELETE /documents/{id}.
func RegisterDeleteRoute(router chi.Router, service *Service) {
	router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		err := service.Delete(request.Context(), strconv.FormatInt(ownerID, 10), chi.URLParam(request, "id"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// RegisterSendRoute mounts POST /documents/{id}/send.
func RegisterSendRoute(router chi.Router, service *Service) {
	router.Post("/{id}/send", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		resp, err := service.Send(request.Context(), strconv.FormatInt(ownerID, 10), chi.URLParam(request, "id"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

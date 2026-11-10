package signing

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// DocumentRoutes returns the router that mounts the certificate and
// audit-trail download endpoints under a document-scoped subrouter.
func DocumentRoutes(service *Service) func(chi.Router) {
	return func(router chi.Router) {
		router.Get("/{docId}/certificate", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := kit.FromContext(request.Context())
			docID := chi.URLParam(request, "docId")

			certPath, err := service.GetOrGenerateCertificate(request.Context(), strconv.FormatInt(identity.UserID, 10), docID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}

			w.Header().Set("Content-Disposition", "attachment; filename=certificate.pdf")
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			http.ServeFile(w, request, certPath)
		})

		router.Get("/{docId}/audit-trail", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := kit.FromContext(request.Context())
			docID := chi.URLParam(request, "docId")

			trailPath, err := service.GetOrGenerateAuditTrail(request.Context(), strconv.FormatInt(identity.UserID, 10), docID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}

			w.Header().Set("Content-Disposition", "attachment; filename=audit_trail.pdf")
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			http.ServeFile(w, request, trailPath)
		})
	}
}

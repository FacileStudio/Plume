package signing

import (
	"net/http"

	"api/internal/authcontext"
	"api/internal/httpjson"

	"github.com/go-chi/chi/v5"
)

func DocumentRoutes(service *Service) func(chi.Router) {
	return func(router chi.Router) {
		router.Get("/{docId}/certificate", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			docID := chi.URLParam(request, "docId")

			certPath, err := service.GetOrGenerateCertificate(request.Context(), identity.UserID, docID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}

			w.Header().Set("Content-Disposition", "attachment; filename=certificate.pdf")
			http.ServeFile(w, request, certPath)
		})

		router.Get("/{docId}/audit-trail", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			docID := chi.URLParam(request, "docId")

			trailPath, err := service.GetOrGenerateAuditTrail(request.Context(), identity.UserID, docID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}

			w.Header().Set("Content-Disposition", "attachment; filename=audit_trail.pdf")
			http.ServeFile(w, request, trailPath)
		})
	}
}

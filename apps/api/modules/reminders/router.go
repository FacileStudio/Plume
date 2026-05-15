package reminders

import (
	"net/http"

	"api/internal/authcontext"
	"api/internal/httpjson"

	"github.com/go-chi/chi/v5"
)

func SignerRoutes(service *Service) func(chi.Router) {
	return func(router chi.Router) {
		router.Post("/{id}/remind", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.RemindSigner(request.Context(), identity.UserID, chi.URLParam(request, "id"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	}
}

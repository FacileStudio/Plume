package reminders

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// SignerRoutes returns the router that mounts the send-reminder action
// (POST /{id}/remind) for a signer.
func SignerRoutes(service *Service) func(chi.Router) {
	return func(router chi.Router) {
		router.Post("/{id}/remind", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := kit.FromContext(request.Context())
			resp, err := service.RemindSigner(request.Context(), strconv.FormatInt(identity.UserID, 10), chi.URLParam(request, "id"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	}
}

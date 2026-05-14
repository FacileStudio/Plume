package reminders

import (
	"net/http"

	"api/internal/authcontext"
	"api/internal/httpjson"
	"api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator) {
	router.Route("/signers", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Post("/{id}/remind", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.RemindSigner(request.Context(), identity.UserID, chi.URLParam(request, "id"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
}

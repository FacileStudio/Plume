package fields

import (
	"net/http"

	"api/internal/authcontext"
	"api/internal/httpjson"

	"github.com/go-chi/chi/v5"
)

func DocumentRoutes(service *Service) func(chi.Router) {
	return func(router chi.Router) {
		router.Get("/{docId}/fields", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.List(request.Context(), identity.UserID, chi.URLParam(request, "docId"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/{docId}/fields", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var req CreateFieldRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.Create(request.Context(), identity.UserID, chi.URLParam(request, "docId"), &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Put("/{docId}/fields/{fieldId}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var req UpdateFieldRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.Update(request.Context(), identity.UserID, chi.URLParam(request, "docId"), chi.URLParam(request, "fieldId"), &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Delete("/{docId}/fields/{fieldId}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			err := service.Delete(request.Context(), identity.UserID, chi.URLParam(request, "docId"), chi.URLParam(request, "fieldId"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	}
}

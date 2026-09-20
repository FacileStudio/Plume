package spaces

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterMemberListRoute mounts GET /spaces/{spaceId}/members.
func RegisterMemberListRoute(router chi.Router, service *Service) {
	router.Get("/{spaceId}/members", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid space ID"))
			return
		}

		resp, err := service.ListMembers(request.Context(), ownerID, spaceID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterMemberAddRoute mounts POST /spaces/{spaceId}/members.
func RegisterMemberAddRoute(router chi.Router, service *Service) {
	router.Post("/{spaceId}/members", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid space ID"))
			return
		}

		var req AddMemberRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}

		resp, err := service.AddMember(request.Context(), ownerID, spaceID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusCreated, resp)
	})
}

// RegisterMemberUpdateRoute mounts PUT /spaces/{spaceId}/members/{memberId}.
func RegisterMemberUpdateRoute(router chi.Router, service *Service) {
	router.Put("/{spaceId}/members/{memberId}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid space ID"))
			return
		}
		memberID, err := strconv.ParseInt(chi.URLParam(request, "memberId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid member ID"))
			return
		}

		var req UpdateMemberRoleRequest
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}

		resp, err := service.UpdateMemberRole(request.Context(), ownerID, spaceID, memberID, &req)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// RegisterMemberRemoveRoute mounts DELETE /spaces/{spaceId}/members/{memberId}.
func RegisterMemberRemoveRoute(router chi.Router, service *Service) {
	router.Delete("/{spaceId}/members/{memberId}", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		spaceID, err := strconv.ParseInt(chi.URLParam(request, "spaceId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid space ID"))
			return
		}
		memberID, err := strconv.ParseInt(chi.URLParam(request, "memberId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid member ID"))
			return
		}

		if err := service.RemoveMember(request.Context(), ownerID, spaceID, memberID); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

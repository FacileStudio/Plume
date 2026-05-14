package documents

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"api/internal/authcontext"
	"api/internal/errors"
	"api/internal/httpjson"
	"api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

const maxUploadSize = 50 << 20

func RegisterRoutes(router chi.Router, service *Service, authService middleware.Authenticator, nested ...func(chi.Router)) {
	router.Route("/documents", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		for _, fn := range nested {
			fn(router)
		}

		router.Post("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())

			if err := request.ParseMultipartForm(maxUploadSize); err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid multipart form"))
				return
			}

			name := request.FormValue("name")
			if name == "" {
				httpjson.WriteError(w, errors.Invalid("name is required"))
				return
			}

			file, header, err := request.FormFile("file")
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("file is required"))
				return
			}
			defer file.Close()

			resp, err := service.Create(request.Context(), identity.UserID, name, header.Filename)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}

			ownerDir := filepath.Join(service.uploadDir, identity.UserID)
			if err := os.MkdirAll(ownerDir, 0o755); err != nil {
				httpjson.WriteError(w, errors.Internal("failed to create upload directory", err))
				return
			}

			storedName := fmt.Sprintf("%d_%s", resp.ID, header.Filename)
			fullPath := filepath.Join(ownerDir, storedName)

			dst, err := os.Create(fullPath)
			if err != nil {
				httpjson.WriteError(w, errors.Internal("failed to save file", err))
				return
			}
			defer dst.Close()

			hasher := sha256.New()
			if _, err := io.Copy(io.MultiWriter(dst, hasher), file); err != nil {
				httpjson.WriteError(w, errors.Internal("failed to write file", err))
				return
			}
			originalHash := hex.EncodeToString(hasher.Sum(nil))

			relativePath := filepath.Join(identity.UserID, storedName)
			if err := service.UpdateStorage(request.Context(), resp.ID, relativePath, originalHash); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp.OriginalHash = originalHash

			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Get("/stats", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.Stats(request.Context(), identity.UserID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			status := request.URL.Query().Get("status")
			resp, err := service.List(request.Context(), identity.UserID, status)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Get("/{id}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.Get(request.Context(), identity.UserID, chi.URLParam(request, "id"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Get("/{id}/file", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			filePath, err := service.GetFilePath(request.Context(), identity.UserID, chi.URLParam(request, "id"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			http.ServeFile(w, request, filePath)
		})

		router.Put("/{id}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var req UpdateRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.Update(request.Context(), identity.UserID, chi.URLParam(request, "id"), &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			err := service.Delete(request.Context(), identity.UserID, chi.URLParam(request, "id"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		router.Post("/{id}/send", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.Send(request.Context(), identity.UserID, chi.URLParam(request, "id"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
}

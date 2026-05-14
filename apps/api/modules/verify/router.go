package verify

import (
	"io"
	"net/http"
	"strings"

	"api/internal/errors"
	"api/internal/hashing"
	"api/internal/httpjson"
	"api/internal/middleware"

	"github.com/go-chi/chi/v5"
)

const maxVerifyUploadSize = 50 << 20

func RegisterRoutes(router chi.Router, service *Service, limiter func(http.Handler) http.Handler) {
	if limiter == nil {
		limiter = func(next http.Handler) http.Handler { return next }
	}

	router.Route("/verify", func(router chi.Router) {
		router.Use(limiter)
		router.Use(middleware.NoStore)

		router.Post("/", func(w http.ResponseWriter, request *http.Request) {
			request.Body = http.MaxBytesReader(w, request.Body, maxVerifyUploadSize)

			reader, err := request.MultipartReader()
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("expected multipart upload"))
				return
			}

			var hash string
			for {
				part, err := reader.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("invalid multipart upload"))
					return
				}
				if part.FormName() != "file" {
					part.Close()
					continue
				}
				hash, err = hashing.SHA256Reader(part)
				part.Close()
				if err != nil {
					httpjson.WriteError(w, errors.Invalid("file too large or unreadable"))
					return
				}
				break
			}

			if hash == "" {
				httpjson.WriteError(w, errors.Invalid("file is required"))
				return
			}

			resp, err := service.Lookup(request.Context(), hash)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Get("/{hash}", func(w http.ResponseWriter, request *http.Request) {
			hash := strings.ToLower(chi.URLParam(request, "hash"))
			if !hashing.IsValidHex(hash) {
				httpjson.WriteError(w, errors.Invalid("hash must be a 64-character hex string"))
				return
			}

			resp, err := service.Lookup(request.Context(), hash)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
}

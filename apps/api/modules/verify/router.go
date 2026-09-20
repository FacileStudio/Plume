package verify

import (
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/FacileStudio/Plume/apps/api/internal/hashing"
	"github.com/FacileStudio/Plume/apps/api/internal/middleware"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

const maxVerifyUploadSize = 50 << 20

// RegisterRoutes mounts the public document-verification endpoints: multipart
// upload hashed on the fly and hash lookup.
func RegisterRoutes(router chi.Router, service *Service, limiter func(http.Handler) http.Handler) {
	if limiter == nil {
		limiter = func(next http.Handler) http.Handler { return next }
	}

	router.Route("/verify", func(router chi.Router) {
		router.Use(limiter)
		router.Use(middleware.NoStore)

		RegisterVerifyUploadRoute(router, service)
		RegisterVerifyLookupRoute(router, service)
	})
}

// RegisterVerifyUploadRoute mounts POST /verify, which hashes an uploaded PDF.
func RegisterVerifyUploadRoute(router chi.Router, service *Service) {
	router.Post("/", func(w http.ResponseWriter, request *http.Request) {
		request.Body = http.MaxBytesReader(w, request.Body, maxVerifyUploadSize)

		hash, err := uploadHash(request)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}

		resp, err := service.Lookup(request.Context(), hash)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// filePart walks the multipart body and returns the file part.
func filePart(request *http.Request) (*multipart.Part, error) {
	reader, err := request.MultipartReader()
	if err != nil {
		return nil, errors.Invalid("expected multipart upload")
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			return nil, errors.Invalid("file is required")
		}
		if err != nil {
			return nil, errors.Invalid("invalid multipart upload")
		}
		if part.FormName() == "file" {
			return part, nil
		}
		part.Close()
	}
}

// uploadHash returns the SHA-256 hash of the upload's file part.
func uploadHash(request *http.Request) (string, error) {
	file, err := filePart(request)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash, err := hashing.SHA256Reader(file)
	if err != nil {
		return "", errors.Invalid("file too large or unreadable")
	}
	return hash, nil
}

// RegisterVerifyLookupRoute mounts GET /verify/{hash}.
func RegisterVerifyLookupRoute(router chi.Router, service *Service) {
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
}

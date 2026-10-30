package documents

import (
	"encoding/json"
	"net/http"

	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterSignatureRoutes mounts POST /documents/signature-requests and
// GET /documents/lookup.
func RegisterSignatureRoutes(router chi.Router, service *Service) {
	router.Post("/signature-requests", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		req, parts, err := parseSignatureRequest(w, request)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		defer parts.File.Close()
		resp, created, err := service.CreateSignatureRequest(request.Context(), ownerID, req, parts.FileName, parts.File)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		status := http.StatusOK
		if created {
			status = http.StatusCreated
		}
		httpjson.WriteJSON(w, status, resp)
	})

	router.Get("/lookup", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)
		resp, err := service.LookupByExternalRef(request.Context(), ownerID, request.URL.Query().Get("external_ref"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	})
}

// parseSignatureRequest reads the multipart body: the PDF in `file` and the
// JSON description in `request`.
func parseSignatureRequest(w http.ResponseWriter, request *http.Request) (*SignatureRequest, *uploadParts, error) {
	request.Body = http.MaxBytesReader(w, request.Body, maxUploadSize)
	if err := request.ParseMultipartForm(maxUploadSize); err != nil {
		return nil, nil, errors.Invalid("invalid upload")
	}
	var req SignatureRequest
	if err := json.Unmarshal([]byte(request.FormValue("request")), &req); err != nil {
		return nil, nil, errors.Invalid("request must be a JSON object")
	}
	file, header, err := request.FormFile("file")
	if err != nil {
		return nil, nil, errors.Invalid("file is required")
	}
	if err := ensurePDFHeader(file); err != nil {
		file.Close()
		return nil, nil, err
	}
	return &req, &uploadParts{FileName: header.Filename, File: file}, nil
}

package documents

import (
	"crypto/sha256"
	"encoding/hex"
	stderrors "errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

const maxUploadSize = 50 << 20

// uploadParts is a validated multipart upload: the form fields and the
// opened file part, ready to be stored.
type uploadParts struct {
	Name     string
	FileName string
	SpaceID  string
	File     multipart.File
}

// parseUpload reads and validates a multipart upload: body size, name field
// and PDF magic bytes. The returned parts carry the opened file for storage.
func parseUpload(w http.ResponseWriter, request *http.Request) (*uploadParts, error) {
	request.Body = http.MaxBytesReader(w, request.Body, maxUploadSize)
	if err := request.ParseMultipartForm(maxUploadSize); err != nil {
		if _, ok := stderrors.AsType[*http.MaxBytesError](err); ok {
			return nil, errors.New("resource_exhausted", "file is too large — maximum upload size is 50 MB", nil)
		}
		return nil, errors.Invalid("invalid upload")
	}

	name := request.FormValue("name")
	if name == "" {
		return nil, errors.Invalid("name is required")
	}

	file, header, err := request.FormFile("file")
	if err != nil {
		return nil, errors.Invalid("file is required")
	}

	if err := ensurePDFHeader(file); err != nil {
		file.Close()
		return nil, err
	}

	return &uploadParts{Name: name, FileName: header.Filename, SpaceID: request.FormValue("space_id"), File: file}, nil
}

// ensurePDFHeader rejects an upload that does not start with the PDF magic
// bytes, rewinding the file when it does.
func ensurePDFHeader(file multipart.File) error {
	pdfHeader := make([]byte, 5)
	if _, err := io.ReadFull(file, pdfHeader); err != nil || string(pdfHeader) != "%PDF-" {
		return errors.Invalid("uploaded file is not a valid PDF")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return errors.Internal("failed to read uploaded file", err)
	}
	return nil
}

// storeUpload writes parts.File under the owner's upload directory, hashes
// it, records the storage path on the document, and returns the file hash.
func storeUpload(request *http.Request, service *Service, parts *uploadParts, ownerID, docID int64) (string, error) {
	ownerDir := filepath.Join(service.uploadDir, strconv.FormatInt(ownerID, 10))
	if err := os.MkdirAll(ownerDir, 0o755); err != nil {
		return "", errors.Internal("failed to create upload directory", err)
	}

	storedName := storedFileName(docID, parts.FileName)
	fullPath := filepath.Join(ownerDir, storedName)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", errors.Internal("failed to save file", err)
	}
	defer dst.Close()

	hash, err := copyAndHash(dst, parts.File)
	if err != nil {
		return "", errors.Internal("failed to write file", err)
	}

	relativePath := filepath.Join(strconv.FormatInt(ownerID, 10), storedName)
	if err := service.UpdateStorage(request.Context(), docID, relativePath, hash); err != nil {
		return "", err
	}
	return hash, nil
}

// storedFileName prefixes the original file name with the document ID so that
// concurrent uploads never collide inside the owner's directory.
func storedFileName(docID int64, fileName string) string {
	return fmt.Sprintf("%d_%s", docID, filepath.Base(fileName))
}

// copyAndHash streams src into dst while computing its SHA-256.
func copyAndHash(dst io.Writer, src io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(io.MultiWriter(dst, hasher), src); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// RegisterUploadRoute mounts POST /documents, the multipart upload route.
func RegisterUploadRoute(router chi.Router, service *Service) {
	router.Post("/", func(w http.ResponseWriter, request *http.Request) {
		ownerID, _ := ownerIDFrom(request)

		parts, err := parseUpload(w, request)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		defer parts.File.Close()

		resp, err := service.Create(request.Context(), strconv.FormatInt(ownerID, 10), parts.Name, parts.FileName, parts.SpaceID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}

		hash, err := storeUpload(request, service, parts, ownerID, resp.ID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		resp.OriginalHash = hash

		httpjson.WriteJSON(w, http.StatusCreated, resp)
	})
}

package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/modules/auth"
	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/testdb"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

const (
	contractOriginalHash = "1111111111111111111111111111111111111111111111111111111111111111"
	contractSignedHash   = "2222222222222222222222222222222222222222222222222222222222222222"
	verifyRefusalCode    = "rate_limited"
)

// contractRouter is the production router over a test database seeded with
// one completed document whose external signer reaches it by token.
func contractRouter(t *testing.T) chi.Router {
	t.Helper()
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_contract_test", Migrate: schemas.Migrate})
	if err != nil {
		t.Fatalf("test database: %v", err)
	}
	seedContract(t, db)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	appEnv := env.Config{}
	sessions, err := session.New(appEnv.Porte(), session.Deps{Sessions: portepg.New(sqlDB).Sessions(), Logger: logger})
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}
	kit, err := oidc.New(context.Background(), appEnv.Porte(), oidc.Deps{Sessions: sessions, Logger: logger})
	if err != nil {
		t.Fatalf("oidc.New: %v", err)
	}
	users, err := auth.NewUsers(context.Background(), sqlDB, logger)
	if err != nil {
		t.Fatalf("NewUsers: %v", err)
	}
	router, err := buildRouter(newServices(db, appEnv, logger, authParts{sessions: sessions, users: users}), sqlDB, appEnv, logger, kit)
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	return router
}

func seedContract(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`TRUNCATE users, documents, signers, fields RESTART IDENTITY CASCADE`,
		`INSERT INTO users (id, email, name, created_at) VALUES (1, 'owner@facile.studio', 'Owner', '2026-01-02T03:04:05Z')`,
		`INSERT INTO documents (id, name, status, file_name, original_hash, signed_hash, owner_id, created_at, updated_at)
		 VALUES (1, 'Contract', 'completed', 'contract.pdf', '` + contractOriginalHash + `', '` + contractSignedHash + `', 1,
		 '2026-01-02T03:04:05Z', '2026-01-03T04:05:06Z')`,
		`INSERT INTO signers (id, document_id, name, email, role, status, token, order_num, signed_at, created_at)
		 VALUES (1, 1, 'External Signer', 'external@example.com', 'signer', 'signed', 'contract-token', 0,
		 '2026-01-03T04:05:06Z', '2026-01-02T03:04:05Z')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

func contractCall(router http.Handler, method, path, remoteAddr string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	request.RemoteAddr = remoteAddr
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// localStamp is how the API renders a January 2026 UTC instant: GORM reads
// timestamptz in the process's local zone.
func localStamp(day, hour, minute, second int) string {
	return time.Date(2026, time.January, day, hour, minute, second, 0, time.UTC).Local().Format(time.RFC3339)
}

func TestVerifyAnswersAreAContract(t *testing.T) {
	router := contractRouter(t)
	created, completed := localStamp(2, 3, 4, 5), localStamp(3, 4, 5, 6)
	signed := `{"match":true,"hash":"` + contractSignedHash + `","variant":"signed","document":{"name":"Contract",` +
		`"file_name":"contract.pdf","status":"completed","created_at":"` + created + `",` +
		`"completed_at":"` + completed + `"},"signers":[{"name":"External Signer","email":"e*******@e**********",` +
		`"status":"signed","signed_at":"` + completed + `"}]}`
	cases := []struct {
		method, path string
		status       int
		body         string
	}{
		{http.MethodGet, "/api/verify/" + contractSignedHash, 200, signed},
		{http.MethodGet, "/api/verify/" + strings.ToUpper(contractOriginalHash), 200, strings.Replace(strings.Replace(signed,
			contractSignedHash, contractOriginalHash, 1), `"variant":"signed"`, `"variant":"original"`, 1)},
		{http.MethodGet, "/api/verify/" + strings.Repeat("0", 64), 200, `{"match":false,"hash":"` + strings.Repeat("0", 64) + `"}`},
		{http.MethodGet, "/api/verify/xyz", 400, `{"error":{"code":"invalid_argument","message":"hash must be a 64-character hex string"}}`},
		{http.MethodPost, "/api/verify", 400, `{"error":{"code":"invalid_argument","message":"expected multipart upload"}}`},
	}
	for i, c := range cases {
		got := contractCall(router, c.method, c.path, "192.0.2."+strconv.Itoa(i+1)+":1000")
		if got.Code != c.status || strings.TrimSpace(got.Body.String()) != c.body {
			t.Errorf("%s %s = %d %s, want %d %s", c.method, c.path, got.Code, got.Body.String(), c.status, c.body)
		}
		if got.Header().Get("Cache-Control") != "no-store, max-age=0" || got.Header().Get("Referrer-Policy") != "no-referrer" {
			t.Errorf("%s %s headers = %v", c.method, c.path, got.Header())
		}
	}
}

func TestVerifyRateLimitIsAContract(t *testing.T) {
	router := contractRouter(t)
	path := "/api/verify/" + contractSignedHash
	for i := range 10 {
		if got := contractCall(router, http.MethodGet, path, "198.51.100.7:"+strconv.Itoa(2000+i)); got.Code != http.StatusOK {
			t.Fatalf("request %d within the burst = %d", i+1, got.Code)
		}
	}
	refused := contractCall(router, http.MethodGet, path, "198.51.100.7:3000")
	if refused.Code != http.StatusTooManyRequests {
		t.Fatalf("request 11 from one host = %d, want 429", refused.Code)
	}
	var body struct {
		Error struct{ Code, Message string } `json:"error"`
	}
	if err := json.Unmarshal(refused.Body.Bytes(), &body); err != nil || body.Error.Code != verifyRefusalCode || body.Error.Message != "too many requests" {
		t.Errorf("refusal body = %s", refused.Body.String())
	}
	if refused.Header().Get("Content-Type") != "application/json" {
		t.Errorf("refusal content type = %q", refused.Header().Get("Content-Type"))
	}
	if wait, err := strconv.Atoi(refused.Header().Get("Retry-After")); err != nil || wait < 1 || wait > 60 {
		t.Errorf("Retry-After = %q", refused.Header().Get("Retry-After"))
	}
	if got := contractCall(router, http.MethodGet, path, "198.51.100.8:2000"); got.Code != http.StatusOK {
		t.Errorf("another host shares the bucket: %d", got.Code)
	}
	if got := contractCall(router, http.MethodGet, "/api/sign/contract-token/status", "198.51.100.7:4000"); got.Code != http.StatusOK {
		t.Errorf("the verify limit reaches the signing routes: %d", got.Code)
	}
}

func TestSigningLinkAnswersAreAContract(t *testing.T) {
	router := contractRouter(t)
	closed := `{"error":{"code":"invalid_argument","message":"document is not available for signing"}}`
	cases := []struct {
		method, path string
		status       int
		body         string
	}{
		{http.MethodGet, "/api/sign/unknown", 404, `{"error":{"code":"not_found","message":"invalid signing link"}}`},
		{http.MethodGet, "/api/sign/contract-token", 400, closed},
		{http.MethodPost, "/api/sign/contract-token/decline", 400, `{"error":{"code":"invalid_argument","message":"already signed or declined"}}`},
		{http.MethodGet, "/api/sign/contract-token/status", 200, ""},
	}
	for _, c := range cases {
		got := contractCall(router, c.method, c.path, "203.0.113.1:1000")
		if got.Code != c.status || (c.body != "" && strings.TrimSpace(got.Body.String()) != c.body) {
			t.Errorf("%s %s = %d %s, want %d %s", c.method, c.path, got.Code, got.Body.String(), c.status, c.body)
		}
	}
}

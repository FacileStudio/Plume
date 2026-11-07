package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/testdb"
)

func TestSigningLinksStaySessionlessWhileTheRestNeedsASession(t *testing.T) {
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_main_test", Migrate: schemas.Migrate})
	if err != nil {
		t.Fatalf("test database: %v", err)
	}
	if err := db.Exec(`TRUNCATE users, documents, signers, fields RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear: %v", err)
	}
	owner := schemas.User{Email: "owner@facile.studio", Name: "Owner"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatalf("seed owner: %v", err)
	}
	doc := schemas.Document{Name: "Contract", OwnerID: owner.ID, Status: "pending"}
	if err := db.Create(&doc).Error; err != nil {
		t.Fatalf("seed document: %v", err)
	}
	signer := schemas.Signer{DocumentID: doc.ID, Name: "External", Email: "external@example.com", Token: "public-token", Status: "pending"}
	if err := db.Create(&signer).Error; err != nil {
		t.Fatalf("seed signer: %v", err)
	}

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
	router := buildRouter(newServices(db, appEnv, logger, sessions), sqlDB, appEnv, logger, kit)

	call := func(method, path, body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		return recorder
	}

	for _, path := range []string{"/api/sign/public-token", "/api/sign/public-token/status", "/api/sign/public-token/opened.gif"} {
		if got := call(http.MethodGet, path, ""); got.Code != http.StatusOK {
			t.Errorf("GET %s without a session = %d, want 200: %s", path, got.Code, got.Body.String())
		}
	}
	if got := call(http.MethodGet, "/api/sign/unknown-token", ""); got.Code != http.StatusNotFound {
		t.Errorf("GET with an unknown token = %d, want 404", got.Code)
	}
	if got := call(http.MethodPost, "/api/sign/public-token/decline", ""); got.Code != http.StatusOK {
		t.Errorf("POST decline without a session = %d, want 200: %s", got.Code, got.Body.String())
	}

	for _, path := range []string{"/api/documents", "/api/auth/me", "/api/spaces", "/api/clients"} {
		if got := call(http.MethodGet, path, ""); got.Code != http.StatusUnauthorized {
			t.Errorf("GET %s without a session = %d, want 401", path, got.Code)
		}
	}
	for _, path := range []string{"/api/auth/register", "/api/auth/login", "/api/auth/password"} {
		if got := call(http.MethodPost, path, `{"email":"a@b.c","password":"correct horse battery"}`); got.Code != http.StatusNotFound && got.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST %s = %d, want it gone", path, got.Code)
		}
	}
}

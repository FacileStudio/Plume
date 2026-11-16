package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/internal/secrets"
	"github.com/FacileStudio/Plume/apps/api/internal/smtptest"
	"github.com/FacileStudio/Plume/apps/api/modules/auth"
	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/testdb"
	"github.com/go-chi/chi/v5"
	"github.com/go-pdf/fpdf"
	"gorm.io/gorm"
)

var shareLink = regexp.MustCompile(`/share/([0-9a-f]{64})`)

type atRest struct {
	t      *testing.T
	db     *gorm.DB
	svc    *services
	router chi.Router
	owner  schemas.User
	mail   *smtptest.Server
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbURL, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	db, err := testdb.Open(dbURL, testdb.Config{Prefix: "plume_main_test", Migrate: schemas.Migrate})
	if err != nil {
		t.Fatalf("test database: %v", err)
	}
	if err := db.Exec(`TRUNCATE users, documents, signers, fields, smtp_configs, webhooks RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear: %v", err)
	}
	return db
}

func buildTestRouter(t *testing.T, db *gorm.DB, appEnv env.Config, sealer *portepg.TokenSealer) (*services, chi.Router) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sessions, err := session.New(appEnv.Porte(), session.Deps{Sessions: portepg.New(sqlDB).Sessions(), Logger: logger})
	if err != nil {
		t.Fatal(err)
	}
	kit, err := oidc.New(context.Background(), appEnv.Porte(), oidc.Deps{Sessions: sessions, Logger: logger})
	if err != nil {
		t.Fatal(err)
	}
	users, err := auth.NewUsers(context.Background(), sqlDB, logger)
	if err != nil {
		t.Fatal(err)
	}
	svc := newServices(db, appEnv, logger, authParts{sessions: sessions, users: users, sealer: sealer})
	router, err := buildRouter(svc, sqlDB, appEnv, logger, kit)
	if err != nil {
		t.Fatal(err)
	}
	return svc, router
}

func seedOwnerWithMail(t *testing.T, db *gorm.DB, sealer *portepg.TokenSealer, mail *smtptest.Server) schemas.User {
	t.Helper()
	owner := schemas.User{Email: "owner@facile.studio", Name: "Owner"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	password, err := sealer.Seal("smtp-pass")
	if err != nil {
		t.Fatal(err)
	}
	config := schemas.SmtpConfig{OwnerID: owner.ID, Host: mail.Host(), Port: mail.Port(), Username: "mailer", Password: password, FromEmail: "from@facile.studio"}
	if err := db.Create(&config).Error; err != nil {
		t.Fatal(err)
	}
	return owner
}

func newAtRest(t *testing.T, sealer *portepg.TokenSealer) *atRest {
	t.Helper()
	db := openTestDB(t)
	appEnv := env.Config{Domain: "http://plume.test", UploadDir: t.TempDir(), SMTPAllowPrivateHosts: true}
	svc, router := buildTestRouter(t, db, appEnv, sealer)
	mail := smtptest.Start(t)
	return &atRest{t: t, db: db, svc: svc, router: router, mail: mail, owner: seedOwnerWithMail(t, db, sealer, mail)}
}

func (a *atRest) request(method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	a.router.ServeHTTP(recorder, request)
	return recorder
}

func (a *atRest) sendDocument() (documents.SignatureSignerStatus, string) {
	a.t.Helper()
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	var file bytes.Buffer
	if err := pdf.Output(&file); err != nil {
		a.t.Fatal(err)
	}
	created, _, err := a.svc.documents.CreateSignatureRequest(context.Background(), a.owner.ID, &documents.SignatureRequest{
		Name: "Contract", Signers: []documents.SignatureSigner{{Name: "Ada", Email: "ada@example.test"}},
	}, "contract.pdf", &file)
	if err != nil {
		a.t.Fatalf("create the signature request: %v", err)
	}
	signer := created.Signers[0]
	prefix := "http://plume.test/share/"
	if !strings.HasPrefix(signer.SigningURL, prefix) {
		a.t.Fatalf("the response must keep returning the usable link, got %q", signer.SigningURL)
	}
	return signer, strings.TrimPrefix(signer.SigningURL, prefix)
}

func (a *atRest) row(signerID int64) schemas.Signer {
	a.t.Helper()
	var row schemas.Signer
	if err := a.db.First(&row, signerID).Error; err != nil {
		a.t.Fatal(err)
	}
	return row
}

func (a *atRest) linkFrom(body string) string {
	a.t.Helper()
	match := shareLink.FindStringSubmatch(body)
	if match == nil {
		a.t.Fatalf("the email carries no signing link: %s", body)
	}
	return match[1]
}

func (a *atRest) status(token string) int {
	return a.request(http.MethodGet, "/api/sign/"+url.PathEscape(token), "").Code
}

func (a *atRest) assertStoredToken(signerID int64, token string, sealed bool) {
	a.t.Helper()
	row := a.row(signerID)
	if row.TokenHash != secrets.HashToken(token) {
		a.t.Fatalf("token_hash %q is not the SHA-256 of the real token", row.TokenHash)
	}
	if got := strings.HasPrefix(row.Token, "pv1:"); got != sealed {
		a.t.Fatalf("token column sealed=%v, want %v: %q", got, sealed, row.Token)
	}
	if !sealed && row.Token != token {
		a.t.Fatalf("without a key the column holds the token as it is, got %q", row.Token)
	}
	if sealed && a.status(row.Token) != http.StatusNotFound {
		a.t.Fatal("the sealed column value must not open the signing view")
	}
}

func (a *atRest) assertEmailAndOwnerLink(row schemas.Signer, token string) {
	a.t.Helper()
	first := a.mail.WaitBodies(a.t, 1)
	if got := a.linkFrom(first[0]); got != token {
		a.t.Fatalf("the signing email carries %q, want the real token", got)
	}
	if got := a.mail.Passwords(); len(got) != 1 || got[0] != "smtp-pass" {
		a.t.Fatalf("the SMTP server received %v, want the opened password", got)
	}
	if a.status(token) != http.StatusOK {
		a.t.Fatal("the real link must open the signing view")
	}
	listed, err := a.svc.signers.ListSigners(context.Background(), strconv.FormatInt(a.owner.ID, 10), strconv.FormatInt(row.DocumentID, 10))
	if err != nil || listed[0].Token != token {
		a.t.Fatalf("the owner's copy-link must stay usable: %v", err)
	}
}

func TestSigningLinkLifecycleWithAndWithoutAKey(t *testing.T) {
	sealer, err := portepg.ParseTokenSealer(strings.Repeat("ab", 32))
	if err != nil {
		t.Fatal(err)
	}
	for name, key := range map[string]*portepg.TokenSealer{"key set": sealer, "no key": nil} {
		t.Run(name, func(t *testing.T) {
			stack := newAtRest(t, key)
			signer, token := stack.sendDocument()
			stack.assertStoredToken(signer.ID, token, key != nil)
			row := stack.row(signer.ID)
			stack.assertEmailAndOwnerLink(row, token)
			stack.remind(row)
			reminded := stack.linkFrom(stack.mail.WaitBodies(t, 2)[1])
			if reminded != token || stack.status(reminded) != http.StatusOK {
				t.Fatalf("the reminder must carry a working link, got %q", reminded)
			}
			stack.sign(row, token)
		})
	}
}

func (a *atRest) remind(row schemas.Signer) {
	a.t.Helper()
	old := time.Now().Add(-48 * time.Hour)
	if err := a.db.Model(&schemas.Signer{}).Where("id = ?", row.ID).Update("last_reminded_at", old).Error; err != nil {
		a.t.Fatal(err)
	}
	if _, err := a.svc.reminders.RemindSigner(context.Background(), strconv.FormatInt(a.owner.ID, 10), strconv.FormatInt(row.ID, 10)); err != nil {
		a.t.Fatalf("remind: %v", err)
	}
}

func (a *atRest) sign(row schemas.Signer, token string) {
	a.t.Helper()
	var field schemas.Field
	if err := a.db.Where("signer_id = ?", row.ID).First(&field).Error; err != nil {
		a.t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"fields": []map[string]any{{"field_id": field.ID, "value": "data:image/png;base64,AAAA"}}})
	if got := a.request(http.MethodPost, "/api/sign/"+token, string(body)); got.Code != http.StatusOK {
		a.t.Fatalf("signing with the real link = %d: %s", got.Code, got.Body.String())
	}
	if got := a.row(row.ID); got.Status != "signed" {
		a.t.Fatalf("signer status %q after signing", got.Status)
	}
	a.mail.WaitBodies(a.t, 3)
}

package mcp_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/modules/auth"
	"github.com/FacileStudio/Plume/apps/api/modules/clients"
	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	"github.com/FacileStudio/Plume/apps/api/modules/fields"
	"github.com/FacileStudio/Plume/apps/api/modules/mcp"
	"github.com/FacileStudio/Plume/apps/api/modules/reminders"
	"github.com/FacileStudio/Plume/apps/api/modules/signers"
	"github.com/FacileStudio/Plume/apps/api/modules/smtp"
	"github.com/FacileStudio/Plume/apps/api/modules/spaces"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/mcpauth"
	"github.com/FacileStudio/tronc/mcp/mcptest"
	"github.com/FacileStudio/tronc/testdb"
	"github.com/go-chi/chi/v5"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	url, ok := testdb.URL()
	if !ok {
		testdb.Announce("sh scripts/postgres-up.sh")
		os.Exit(m.Run())
	}
	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_mcp_test", Migrate: schemas.Migrate})
	if err != nil {
		panic(err)
	}
	testDB = db
	os.Exit(m.Run())
}

// identities resolves each test subject to its user.
type identities struct {
	porte.IdentityStore
	users map[string]int64
}

// Find resolves a known subject.
func (s identities) Find(_ context.Context, provider, sub string) (porte.StoredIdentity, error) {
	userID, ok := s.users[sub]
	if !ok {
		return porte.StoredIdentity{}, porte.ErrNotFound
	}
	return porte.StoredIdentity{UserID: userID, Provider: provider, Subject: sub}, nil
}

// fixture is the MCP endpoint over Plume's real services: an owner, a stranger and a
// shared space holding the owner and a teammate.
type fixture struct {
	db       *gorm.DB
	issuer   *mcptest.Issuer
	resource string
	users    map[string]int64
	space    schemas.Space
	docs     *documents.Service
	front    func(http.Handler) http.Handler
}

// newFixture starts the endpoint and the confirm page routes on any free port.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	return newFixtureOn(t, "")
}

// newFixtureOn starts the endpoint on addr, any free port when empty.
func newFixtureOn(t *testing.T, addr string) *fixture {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	must(t, testDB.Exec(`TRUNCATE users, spaces, space_members, documents, signers, fields, clients, mcp_actions RESTART IDENTITY CASCADE`).Error)
	f := &fixture{db: testDB, issuer: mcptest.NewIssuer(t), users: map[string]int64{}, front: func(next http.Handler) http.Handler { return next }}
	f.seed(t)
	var handler http.Handler
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { f.front(handler).ServeHTTP(w, r) }))
	if addr != "" {
		listener, err := net.Listen("tcp", addr)
		must(t, err)
		srv.Listener = listener
	}
	srv.Start()
	t.Cleanup(srv.Close)
	f.resource = srv.URL + "/mcp"
	service := f.service(t, srv.URL)
	router := chi.NewRouter()
	mcp.RegisterRoutes(router, service)
	router.Route("/api", func(r chi.Router) { mcp.RegisterPendingRoutes(r, service, requireTestSession) })
	handler = router
	return f
}

// service builds the MCP server over the fixture's database and issuer.
func (f *fixture) service(t *testing.T, domain string) *mcp.Service {
	t.Helper()
	verifier, err := mcpauth.Verifier(context.Background(), mcpauth.Config{
		Issuer: f.issuer.URL, Resource: f.resource, ScopePrefix: "plume", Identities: identities{users: f.users},
	})
	must(t, err)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sqlDB, err := f.db.DB()
	must(t, err)
	users, err := auth.NewUsers(context.Background(), sqlDB, logger)
	must(t, err)
	mail, hooks := smtp.NewService(f.db), webhooks.NewService(f.db)
	f.docs = documents.NewService(f.db, mail, hooks, domain, t.TempDir())
	service, err := mcp.NewService(mcp.Deps{
		Documents: f.docs, Signers: signers.NewService(f.db, f.docs, hooks, mail, nil, domain),
		Fields: fields.NewService(f.db, f.docs), Reminders: reminders.NewService(f.db, mail, hooks, domain),
		Clients: clients.NewService(f.db), Spaces: spaces.NewService(f.db), Accounts: users,
	}, mcp.Config{
		Resource: f.resource, Issuer: f.issuer.URL, Version: "test", StateKey: []byte("state-key"),
		Verifier: verifier, Pending: mcptest.NewPending(), PerMinute: 100_000, Logger: logger,
	})
	must(t, err)
	return service
}

// seed creates the people and the shared space.
func (f *fixture) seed(t *testing.T) {
	t.Helper()
	for _, name := range []string{"owner", "teammate", "stranger"} {
		user := schemas.User{Email: name + "@example.test", Name: name}
		must(t, f.db.Create(&user).Error)
		f.users[name] = user.ID
	}
	f.space = schemas.Space{Name: "Studio", OwnerID: f.users["owner"]}
	must(t, f.db.Create(&f.space).Error)
	for _, name := range []string{"owner", "teammate"} {
		must(t, f.db.Create(&schemas.SpaceMember{SpaceID: f.space.ID, UserID: f.users[name], Role: "member"}).Error)
	}
}

// token is a token for the owner with scopes, or both when none is given.
func (f *fixture) token(t testing.TB, scopes ...string) string {
	if len(scopes) == 0 {
		scopes = []string{"plume.read", "plume.write"}
	}
	return f.issuer.Token(t, f.resource, "owner", scopes...)
}

// connect opens a session for subject with both scopes.
func (f *fixture) connect(t *testing.T, subject string) *sdk.ClientSession {
	t.Helper()
	return mcptest.Connect(t, f.resource, f.issuer.Token(t, f.resource, subject, "plume.read", "plume.write"), mcptest.ClientOptions{})
}

// draft stores a draft of the owner's with one signer, through the service the API uses.
func (f *fixture) draft(t testing.TB, name string) *documents.SignatureResponse {
	t.Helper()
	send := false
	created, _, err := f.docs.CreateSignatureRequest(context.Background(), f.users["owner"], &documents.SignatureRequest{
		Name: name, Send: &send, Signers: []documents.SignatureSigner{{Name: "Ada", Email: "ada@example.test"}},
	}, name+".pdf", bytes.NewReader(samplePDF(t, 1)))
	must(t, err)
	return created
}

// sent stores a document out for signature and returns it with its signer.
func (f *fixture) sent(t testing.TB, name string) *documents.SignatureResponse {
	t.Helper()
	created := f.draft(t, name)
	_, err := f.docs.Send(context.Background(), strconv.FormatInt(f.users["owner"], 10), strconv.FormatInt(created.Document.ID, 10))
	must(t, err)
	must(t, f.db.Model(&schemas.Signer{}).Where("document_id = ?", created.Document.ID).Update("last_reminded_at", time.Now().Add(-48*time.Hour)).Error)
	return created
}

// status reads a document's status, or gone once it is deleted.
func (f *fixture) status(id int64) string {
	var doc schemas.Document
	if err := f.db.Where("id = ?", id).First(&doc).Error; err != nil {
		return "gone"
	}
	return doc.Status
}

// call runs a tool and decodes its result, failing on a tool error.
func call(t *testing.T, session *sdk.ClientSession, tool string, args map[string]any, out any) {
	t.Helper()
	result := mcptest.Call(t, session, tool, args)
	if result.IsError {
		t.Fatalf("%s: %s", tool, mcptest.Text(result))
	}
	mcptest.Decode(t, result, out)
}

// refused runs a tool and returns its error text, failing when it worked.
func refused(t *testing.T, session *sdk.ClientSession, tool string, args map[string]any) string {
	t.Helper()
	result := mcptest.Call(t, session, tool, args)
	if !result.IsError {
		t.Fatalf("%s must be refused, got %s", tool, mcptest.Text(result))
	}
	return mcptest.Text(result)
}

// must fails the test on err.
func must(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

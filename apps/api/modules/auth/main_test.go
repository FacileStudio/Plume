package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte/local"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/testdb"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	url, ok := testdb.URL()
	if !ok {
		testdb.Announce("sh scripts/postgres-up.sh")
		os.Exit(m.Run())
	}

	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_auth_test", Migrate: schemas.Migrate})
	if err != nil {
		panic(err)
	}
	testDB = db
	os.Exit(m.Run())
}

// api is the real router over a real Postgres, and it has to be: ChangePassword
// reads the caller's session id out of porte.From(ctx) and writes the rotated
// cookie through the ResponseWriter, so a service-level test holding a bare
// context exercises neither half and passes regardless.
type api struct {
	router  chi.Router
	service *Service
	db      *gorm.DB
}

// newAPI builds the auth module the way main.go does, or skips loudly when
// there is no Postgres to build it against.
func newAPI(t *testing.T) *api {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testDB.Exec(`TRUNCATE users RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear users: %v", err)
	}

	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	store := portepg.New(sqlDB)
	users := NewUserStore(testDB)
	appEnv := env.Config{}

	sessions, err := session.New(appEnv.Porte(), session.Deps{Sessions: store.Sessions(), Logger: logger})
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}
	passwords, err := local.New(local.Config{AllowRegistration: true}, local.Deps{
		Users:      users,
		Identities: store.Identities(),
		Sessions:   sessions,
		Logger:     logger,
		Count:      users.CountUsers,
	})
	if err != nil {
		t.Fatalf("local.New: %v", err)
	}

	service := NewService(testDB, sessions, passwords, logger)
	router := chi.NewRouter()
	RegisterRoutes(router, service, appEnv)
	return &api{router: router, service: service, db: testDB}
}

// do sends one request. A bearer is used rather than the cookie porte also
// sets, because that is the transport this app's client uses.
func (a *api) do(t *testing.T, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode body: %v", err)
		}
		payload = encoded
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	a.router.ServeHTTP(recorder, request)
	return recorder
}

// decode reads the response, failing the test on a status that does not match
// and reporting the error code tronc nests under "error".
func decode(t *testing.T, recorder *httptest.ResponseRecorder, want int) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %d body %q: %v", recorder.Code, recorder.Body.String(), err)
	}
	if recorder.Code != want {
		t.Fatalf("status %d, want %d: %v", recorder.Code, want, body)
	}
	return body
}

// errorCode is the machine-readable half of tronc's error envelope.
func errorCode(body map[string]any) string {
	envelope, ok := body["error"].(map[string]any)
	if !ok {
		return ""
	}
	code, _ := envelope["code"].(string)
	return code
}

// register creates an account through the real route and returns its bearer.
func (a *api) register(t *testing.T, email, password string) (int64, string) {
	t.Helper()
	body := decode(t, a.do(t, http.MethodPost, "/auth/register", "", map[string]string{
		"email": email, "password": password,
	}), http.StatusCreated)
	token, _ := body["token"].(string)
	if token == "" {
		t.Fatalf("register returned no token: %v", body)
	}
	var user schemas.User
	if err := a.db.Where("email = ?", email).First(&user).Error; err != nil {
		t.Fatalf("read the registered user: %v", err)
	}
	return user.ID, token
}

// federated creates the account an SSO-only user has: a row in users with no
// password identity behind it.
func (a *api) federated(t *testing.T, email string) (int64, string) {
	t.Helper()
	user := schemas.User{Email: email, Name: "Federated"}
	if err := a.db.Create(&user).Error; err != nil {
		t.Fatalf("create the federated user: %v", err)
	}
	token, _, err := a.service.Sessions().Issue(context.Background(), user.ID, "")
	if err != nil {
		t.Fatalf("issue a session: %v", err)
	}
	return user.ID, token
}

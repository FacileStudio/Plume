package schemas

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

const testIssuer = "https://porte.test/application/o/plume/"

// seedPrePorte rebuilds the shape production is in before this deploy: the old
// sessions table, a federated identity recorded on the user row, and a local
// password hash in users.password_hash.
//
// The legacy sessions table is created here in SQL because the model is gone.
// That is the point: after this migration it exists only in databases that
// predate it, and the only thing that still has to understand it is AdoptPorte.

func seedPrePorte(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`DROP TABLE IF EXISTS sessions`,
		`CREATE TABLE sessions (token text PRIMARY KEY, user_id bigint NOT NULL, expires_at timestamptz, created_at timestamptz)`,
		`INSERT INTO users (id, email, name, oidc_subject, oidc_access_token, oidc_refresh_token, profile_synced_at, password_hash, created_at)
		 VALUES (1, 'camille@facile.studio', 'Camille', 'sub-1', 'ciphertext', 'ciphertext', now(), '', now())`,
		`INSERT INTO users (id, email, name, oidc_subject, password_hash, created_at)
		 VALUES (2, 'Noah@Facile.Studio', 'Noah', NULL, '$argon2id$fake', now())`,
		`INSERT INTO sessions (token, user_id, expires_at, created_at) VALUES
			('live', 1, now() + interval '10 days', now() - interval '40 days'),
			('dead', 1, now() - interval '1 day', now() - interval '31 days')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("seed: %v\n%s", err, statement)
		}
	}
	t.Cleanup(func() {
		db.Exec(`DROP TABLE IF EXISTS sessions`)
		db.Exec(`DELETE FROM users`)
	})
}

// Nobody may be signed out by this deploy. Both tables store the SHA-256 hex of
// a token and nothing else, which is exactly what porte stores, so the rows
// move and the cookie already in somebody's browser keeps authenticating. The
// carried session is stamped with last_used_at rather than copied: carrying the
// created_at over would put it 40 days into the seven-day idle window and sign
// the user out on the deploy meant to keep them.
func TestAdoptPorteKeepsEverybodySignedIn(t *testing.T) {
	db := requireDB(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	var carried struct {
		UserID     int64
		Label      string
		LastUsedAt time.Time
	}
	if err := db.Raw(`SELECT user_id, label, last_used_at FROM porte_sessions WHERE token_hash = 'live'`).Scan(&carried).Error; err != nil {
		t.Fatalf("read the carried session: %v", err)
	}
	if carried.UserID != 1 || carried.Label != "" {
		t.Fatalf("the browser session did not survive as an unlabelled session: %+v", carried)
	}
	if time.Since(carried.LastUsedAt) > time.Hour {
		t.Fatalf("last_used_at was copied instead of stamped: %v", carried.LastUsedAt)
	}

	var expired int64
	if err := db.Raw(`SELECT count(*) FROM porte_sessions WHERE token_hash = 'dead'`).Scan(&expired).Error; err != nil {
		t.Fatalf("count expired: %v", err)
	}
	if expired != 0 {
		t.Fatal("an already-expired session was carried over")
	}

	var remaining *string
	if err := db.Raw(`SELECT to_regclass('sessions')::text`).Scan(&remaining).Error; err != nil {
		t.Fatalf("check sessions: %v", err)
	}
	if remaining != nil {
		t.Fatal("the legacy sessions table survived")
	}

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt is not idempotent: %v", err)
	}
}

// The password hash moves into the identity row porte/local reads. Without it
// the login form answers "invalid credentials" to a correct password, with the
// hash still sitting in the users table and no error anywhere. The local
// identity is keyed on the account id, which is what porte/local has looked one
// up by since v0.3.0; keying on the address is the mutable key that version
// removed.
func TestAdoptPorteMovesThePasswords(t *testing.T) {
	db := requireDB(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	var identity struct {
		UserID       int64
		PasswordHash string
	}
	err := db.Raw(
		`SELECT user_id, password_hash FROM porte_identities WHERE provider = 'local' AND subject = '2'`,
	).Scan(&identity).Error
	if err != nil {
		t.Fatalf("read the local identity: %v", err)
	}
	if identity.UserID != 2 || identity.PasswordHash != "$argon2id$fake" {
		t.Fatalf("the password did not move: %+v", identity)
	}

	var withoutPassword int64
	if err := db.Raw(`SELECT count(*) FROM porte_identities WHERE provider = 'local' AND user_id = 1`).Scan(&withoutPassword).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if withoutPassword != 0 {
		t.Fatal("an account with no password gained a local identity, which is a login that cannot be used and an account that cannot be registered")
	}
}

// An account that registered through porte v0.2 has an address-keyed identity
// and an empty users.password_hash, because CreateFromPassword never writes
// that column — porte holds the credential. It is therefore outside
// adoptExistingPasswords' filter, and the re-key UPDATE is the only statement
// that can reach it. Without that statement the deploy compiles, boots, and
// answers 401 to every correct password with nothing in the logs.
func TestAdoptPorteRekeysAnAddressKeyedIdentity(t *testing.T) {
	db := requireDB(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt: %v", err)
	}
	statements := []string{
		`UPDATE porte_identities SET subject = 'noah@facile.studio' WHERE provider = 'local' AND user_id = 2`,
		`INSERT INTO users (id, email, name, password_hash, created_at)
		 VALUES (3, 'iris@facile.studio', 'Iris', '', now())`,
		`INSERT INTO porte_identities (user_id, provider, subject, password_hash)
		 VALUES (3, 'local', 'iris@facile.studio', '$argon2id$since')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("rewind to the address key: %v\n%s", err, statement)
		}
	}

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("re-adopt: %v", err)
	}

	var subjects []string
	if err := db.Raw(`SELECT subject FROM porte_identities WHERE provider = 'local' ORDER BY subject`).Scan(&subjects).Error; err != nil {
		t.Fatalf("read the local identities: %v", err)
	}
	if len(subjects) != 2 || subjects[0] != "2" || subjects[1] != "3" {
		t.Fatalf("the local identities were not re-keyed onto the account id: %v", subjects)
	}

	var hash string
	if err := db.Raw(`SELECT password_hash FROM porte_identities WHERE provider = 'local' AND subject = '3'`).Scan(&hash).Error; err != nil {
		t.Fatalf("read the re-keyed hash: %v", err)
	}
	if hash != "$argon2id$since" {
		t.Fatalf("the re-key lost the credential: %q", hash)
	}
}

// The federated identity moves off the user row. Without it porte finds no
// identity, falls back to matching the verified email and relinks on the next
// login — which works, but leans the whole existing user base on the weaker of
// the two matching paths, on the one deploy where nobody would notice. The
// provider tokens are deliberately not carried across: Plume encrypts them
// with ENCRYPTION_KEY and porte stores them as it will send them, so handing
// porte a refresh token that is not one makes the first profile sync fail and
// look like the provider revoked it.
func TestAdoptPorteMovesTheOIDCSubject(t *testing.T) {
	db := requireDB(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	var identity struct {
		UserID       int64
		AccessToken  string
		RefreshToken string
	}
	err := db.Raw(
		`SELECT user_id, access_token, refresh_token FROM porte_identities WHERE provider = ? AND subject = 'sub-1'`,
		testIssuer,
	).Scan(&identity).Error
	if err != nil {
		t.Fatalf("read the identity: %v", err)
	}
	if identity.UserID != 1 {
		t.Fatal("the oidc subject was not adopted")
	}
	if identity.AccessToken != "" || identity.RefreshToken != "" {
		t.Fatalf("encrypted provider tokens were carried across: %+v", identity)
	}

	var rows int64
	if err := db.Raw(`SELECT count(*) FROM porte_identities WHERE provider = ?`, testIssuer).Scan(&rows).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if rows != 1 {
		t.Fatalf("expected exactly one federated identity, got %d", rows)
	}
}

// An empty issuer is a deployment with SSO switched off. The sessions and the
// passwords still have to move — they are what keeps people
// signed in and able to sign in — but there is no provider to key a federated
// identity against.
func TestAdoptPorteWithoutAnIssuerStillMovesTheCredentials(t *testing.T) {
	db := requireDB(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, ""); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	var sessions, federated int64
	if err := db.Raw(`SELECT count(*) FROM porte_sessions`).Scan(&sessions).Error; err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if err := db.Raw(`SELECT count(*) FROM porte_identities WHERE provider <> 'local'`).Scan(&federated).Error; err != nil {
		t.Fatalf("count identities: %v", err)
	}
	if sessions != 1 {
		t.Fatalf("expected the one live session, got %d", sessions)
	}
	if federated != 0 {
		t.Fatalf("an identity was keyed against no provider: %d rows", federated)
	}
}

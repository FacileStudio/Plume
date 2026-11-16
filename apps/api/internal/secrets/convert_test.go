package secrets

import (
	"context"
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"

	"gorm.io/gorm"
)

func seedLegacy(t *testing.T, db *gorm.DB) (signer schemas.Signer, smtpRow schemas.SmtpConfig, hook schemas.Webhook) {
	t.Helper()
	signer = schemas.Signer{Name: "A", Email: "a@x.test", Token: "legacyplaintexttoken"}
	smtpRow = schemas.SmtpConfig{OwnerID: 1, Host: "h", Port: 25, Password: "smtp-secret"}
	hook = schemas.Webhook{OwnerID: 1, URL: "https://example.test", Secret: "whsec_legacy", Enabled: true}
	for _, row := range []any{&signer, &smtpRow, &hook} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	return signer, smtpRow, hook
}

func assertOpensTo(t *testing.T, stored, want string) {
	t.Helper()
	if !Sealed(stored) {
		t.Fatalf("not sealed after conversion: %q", stored)
	}
	if opened, err := sealerOf(t, "ab").Open(stored); err != nil || opened != want {
		t.Fatalf("sealed value does not open back to %q: %q %v", want, opened, err)
	}
}

func TestConvertSealsAndHashesLegacyRowsOnceOnly(t *testing.T) {
	db := requireDB(t)
	sealer := sealerOf(t, "ab")
	signer, smtpRow, hook := seedLegacy(t, db)

	counts, err := Convert(context.Background(), db, sealer, quiet())
	if err != nil || counts != (Counts{Signers: 1, SMTP: 1, Webhooks: 1}) {
		t.Fatalf("first pass counts %+v err %v", counts, err)
	}
	token := column(t, db, "signers", "token", signer.ID)
	password := column(t, db, "smtp_configs", "password", smtpRow.ID)
	assertOpensTo(t, token, "legacyplaintexttoken")
	assertOpensTo(t, password, "smtp-secret")
	assertOpensTo(t, column(t, db, "webhooks", "secret", hook.ID), "whsec_legacy")
	if got := column(t, db, "signers", "token_hash", signer.ID); got != HashToken("legacyplaintexttoken") {
		t.Fatalf("token_hash %q", got)
	}
	if found, err := FindSigner(db, "legacyplaintexttoken"); err != nil || found.ID != signer.ID {
		t.Fatalf("the real token must open the converted row: %v", err)
	}
	wantNotFound(t, db, token, "the sealed value must not open the converted row")

	if again, err := Convert(context.Background(), db, sealer, quiet()); err != nil || again != (Counts{}) {
		t.Fatalf("a second pass must change nothing, got %+v %v", again, err)
	}
	if column(t, db, "signers", "token", signer.ID) != token || column(t, db, "smtp_configs", "password", smtpRow.ID) != password {
		t.Fatal("a second pass rewrote sealed values")
	}
}

func TestConvertWithoutAKeyOnlyHashes(t *testing.T) {
	db := requireDB(t)
	signer, smtpRow, hook := seedLegacy(t, db)

	counts, err := Convert(context.Background(), db, nil, quiet())
	if err != nil {
		t.Fatal(err)
	}
	if counts.Signers != 1 || counts.SMTP != 0 || counts.Webhooks != 0 {
		t.Fatalf("keyless counts %+v", counts)
	}
	if column(t, db, "signers", "token", signer.ID) != "legacyplaintexttoken" || column(t, db, "smtp_configs", "password", smtpRow.ID) != "smtp-secret" || column(t, db, "webhooks", "secret", hook.ID) != "whsec_legacy" {
		t.Fatal("without a key the values must stay in clear")
	}
	if got := column(t, db, "signers", "token_hash", signer.ID); got != HashToken("legacyplaintexttoken") {
		t.Fatalf("token_hash %q", got)
	}
	if again, _ := Convert(context.Background(), db, nil, quiet()); again != (Counts{}) {
		t.Fatalf("a second keyless pass must change nothing, got %+v", again)
	}
}

func TestConvertSealsPlaintextRowsThatAlreadyHaveAHash(t *testing.T) {
	db := requireDB(t)
	row := schemas.Signer{Name: "A", Email: "a@x.test", Token: "keylesstoken", TokenHash: HashToken("keylesstoken")}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	sealer := sealerOf(t, "ab")
	if counts, err := Convert(context.Background(), db, sealer, quiet()); err != nil || counts.Signers != 1 {
		t.Fatalf("counts %+v err %v", counts, err)
	}
	if !Sealed(column(t, db, "signers", "token", row.ID)) {
		t.Fatal("a hashed plaintext token must be sealed once a key is set")
	}
}

func TestConvertLeavesAValueNoKeyOpens(t *testing.T) {
	db := requireDB(t)
	foreign, err := sealerOf(t, "cd").Seal("othertoken")
	if err != nil {
		t.Fatal(err)
	}
	row := schemas.Signer{Name: "A", Email: "a@x.test", Token: foreign}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	counts, err := Convert(context.Background(), db, sealerOf(t, "ab"), quiet())
	if err != nil {
		t.Fatal(err)
	}
	if counts.Unreadable != 1 || counts.Signers != 0 {
		t.Fatalf("counts %+v", counts)
	}
	if column(t, db, "signers", "token", row.ID) != foreign {
		t.Fatal("an unreadable value must be left as it is")
	}
}

func TestTwoReplicasConvertingAtOnceChangeEachRowOnce(t *testing.T) {
	db := requireDB(t)
	sealer := sealerOf(t, "ab")
	for i := range 20 {
		row := schemas.Signer{Name: "A", Email: "a@x.test", Token: strings.Repeat("a", 10) + string(rune('a'+i))}
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	results := make(chan Counts, 2)
	for range 2 {
		go func() {
			counts, err := Convert(context.Background(), db, sealer, quiet())
			if err != nil {
				t.Error(err)
			}
			results <- counts
		}()
	}
	total := (<-results).Signers + (<-results).Signers
	if total != 20 {
		t.Fatalf("20 rows converted %d times in total", total)
	}
}

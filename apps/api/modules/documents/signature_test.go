package documents

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/modules/smtp"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/testdb"

	"github.com/go-pdf/fpdf"
)

func twoPagePDF(t *testing.T) []byte {
	t.Helper()
	pdf := fpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.AddPage()
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		t.Fatalf("render pdf: %v", err)
	}
	return buf.Bytes()
}

func signatureService(t *testing.T) (*Service, int64) {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testDB.Exec(`TRUNCATE users, spaces, space_members, documents, signers, fields, clients RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear tables: %v", err)
	}
	svc := NewService(testDB, smtp.NewService(testDB), webhooks.NewService(testDB), "http://plume.test", t.TempDir())
	return svc, createUser(t, "ardoise@facile.studio")
}

func validRequest() *SignatureRequest {
	return &SignatureRequest{
		Name: "Contract 42",
		Link: Link{ExternalRef: "ardoise:contract:42", NuageSpaceFacileID: "fac_0123456789abcdef01234567", NuageLinkedTo: "ardoise:contract:42:signed"},
		Signers: []SignatureSigner{
			{Name: "Ada", Email: "ada@example.com"},
			{Name: "Bob", Email: "bob@example.com", Role: "viewer"},
		},
	}
}

func TestSignatureRequestValidation(t *testing.T) {
	mutations := map[string]func(*SignatureRequest){
		"no name":    func(r *SignatureRequest) { r.Name = "" },
		"no signers": func(r *SignatureRequest) { r.Signers = nil },
		"bad email":  func(r *SignatureRequest) { r.Signers[0].Email = "nope" },
		"bad role":   func(r *SignatureRequest) { r.Signers[0].Role = "boss" },
		"bad ref":    func(r *SignatureRequest) { r.ExternalRef = "a\nb" },
		"bad space":  func(r *SignatureRequest) { r.NuageSpaceFacileID = "fac_x" },
		"bad client": func(r *SignatureRequest) { r.Client = &SignatureClient{Name: "Acme", Email: "nope"} },
		"field off signer": func(r *SignatureRequest) {
			r.Fields = []SignatureField{{Signer: 5, FieldType: "signature", Page: 1, X: 1, Y: 1, Width: 5, Height: 5}}
		},
		"field off page": func(r *SignatureRequest) {
			r.Fields = []SignatureField{{Signer: 0, FieldType: "signature", Page: 1, X: 99, Y: 1, Width: 5, Height: 101}}
		},
		"field type": func(r *SignatureRequest) {
			r.Fields = []SignatureField{{Signer: 0, FieldType: "stamp", Page: 1, X: 1, Y: 1, Width: 5, Height: 5}}
		},
	}
	if err := validRequest().validate(); err != nil {
		t.Fatalf("valid request refused: %v", err)
	}
	for name, mutate := range mutations {
		req := validRequest()
		mutate(req)
		if err := req.validate(); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestCreateSignatureRequestEndToEnd(t *testing.T) {
	svc, owner := signatureService(t)
	ctx := context.Background()

	resp, created, err := svc.CreateSignatureRequest(ctx, owner, validRequest(), "contract.pdf", bytes.NewReader(twoPagePDF(t)))
	if err != nil || !created {
		t.Fatalf("create = %v, created %v", err, created)
	}
	doc := resp.Document
	if doc.Status != "pending" || doc.ExternalRef != "ardoise:contract:42" || doc.NuageLinkedTo != "ardoise:contract:42:signed" {
		t.Fatalf("unexpected document %+v", doc)
	}
	if len(resp.Signers) != 2 || !strings.HasPrefix(resp.Signers[0].SigningURL, "http://plume.test/share/") {
		t.Fatalf("unexpected signers %+v", resp.Signers)
	}
	var fields []schemas.Field
	testDB.Where("document_id = ?", doc.ID).Find(&fields)
	if len(fields) != 1 || fields[0].Page != 2 || fields[0].FieldType != "signature" || !fields[0].Required {
		t.Fatalf("want one required signature box on the last page for the signer only, got %+v", fields)
	}
}

func TestCreateSignatureRequestIsIdempotentOnExternalRef(t *testing.T) {
	svc, owner := signatureService(t)
	ctx := context.Background()
	first, _, err := svc.CreateSignatureRequest(ctx, owner, validRequest(), "c.pdf", bytes.NewReader(twoPagePDF(t)))
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, created, err := svc.CreateSignatureRequest(ctx, owner, validRequest(), "c.pdf", bytes.NewReader(twoPagePDF(t)))
	if err != nil || created || second.Document.ID != first.Document.ID {
		t.Fatalf("second = %+v created=%v err=%v", second, created, err)
	}
	var count int64
	testDB.Model(&schemas.Document{}).Count(&count)
	if count != 1 {
		t.Fatalf("want 1 document, got %d", count)
	}
}

func TestCreateSignatureRequestDiscardsOnBadPDF(t *testing.T) {
	svc, owner := signatureService(t)
	_, _, err := svc.CreateSignatureRequest(context.Background(), owner, validRequest(), "c.pdf", strings.NewReader("%PDF-garbage"))
	if err == nil {
		t.Fatal("garbage PDF accepted")
	}
	var docs, signers int64
	testDB.Model(&schemas.Document{}).Count(&docs)
	testDB.Model(&schemas.Signer{}).Count(&signers)
	if docs != 0 || signers != 0 {
		t.Fatalf("leftovers: %d documents, %d signers", docs, signers)
	}
	entries, _ := os.ReadDir(filepath.Join(svc.uploadDir, "1"))
	if len(entries) != 0 {
		t.Fatalf("stored file not removed: %v", entries)
	}
}

func TestLookupByExternalRefIsScopedToTheOwner(t *testing.T) {
	svc, owner := signatureService(t)
	ctx := context.Background()
	other := createUser(t, "other@facile.studio")
	if _, _, err := svc.CreateSignatureRequest(ctx, owner, validRequest(), "c.pdf", bytes.NewReader(twoPagePDF(t))); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := svc.LookupByExternalRef(ctx, owner, "ardoise:contract:42")
	if err != nil || got.Document.Status != "pending" || len(got.Signers) != 2 {
		t.Fatalf("lookup = %+v, %v", got, err)
	}
	if _, err := svc.LookupByExternalRef(ctx, other, "ardoise:contract:42"); codeOf(err) != "not_found" {
		t.Fatalf("other owner: %v", err)
	}
	if _, err := svc.LookupByExternalRef(ctx, owner, ""); codeOf(err) != "invalid_argument" {
		t.Fatalf("empty ref: %v", err)
	}
}

func TestSignatureRequestFilesTheDocumentUnderTheClient(t *testing.T) {
	svc, owner := signatureService(t)
	ctx := context.Background()
	req := validRequest()
	req.Client = &SignatureClient{Name: "Acme", Email: "Contact@Acme.example", Company: "Acme SAS"}
	first, _, err := svc.CreateSignatureRequest(ctx, owner, req, "c.pdf", bytes.NewReader(twoPagePDF(t)))
	if err != nil || first.Document.ClientID == nil {
		t.Fatalf("first = %+v err=%v", first, err)
	}
	again := validRequest()
	again.ExternalRef = "ardoise:contract:43"
	again.Client = &SignatureClient{Name: "Acme renamed", Email: "contact@acme.example"}
	second, _, err := svc.CreateSignatureRequest(ctx, owner, again, "c.pdf", bytes.NewReader(twoPagePDF(t)))
	if err != nil || second.Document.ClientID == nil || *second.Document.ClientID != *first.Document.ClientID {
		t.Fatalf("same email must reuse the client: %+v err=%v", second, err)
	}
	var clients int64
	testDB.Model(&schemas.Client{}).Count(&clients)
	if clients != 1 {
		t.Fatalf("want 1 client, got %d", clients)
	}
}

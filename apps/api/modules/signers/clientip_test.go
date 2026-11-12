package signers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	"github.com/FacileStudio/Plume/apps/api/modules/smtp"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
	"github.com/FacileStudio/Plume/apps/api/schemas"
	troncenv "github.com/FacileStudio/tronc/env"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// The audit row is evidence, so its IP must be the visitor tronc's real-IP
// middleware resolved, through the production chain: Cloudflare, then Traefik,
// which replaces X-Forwarded-For with the edge and sets X-Real-Ip to it.
func TestDeclineRecordsTheResolvedClientIP(t *testing.T) {
	db := requireDB(t)
	router := productionChain(t, db)

	t.Run("visitor through a cloudflare edge and traefik", func(t *testing.T) {
		got := declineFrom(t, db, router, crossing{token: "tok-cf", peer: "172.18.0.5:41234", edge: "104.16.0.1", visitor: "203.0.113.7"})
		if got != "203.0.113.7" {
			t.Fatalf("audit ip = %q, want the Cf-Connecting-Ip visitor", got)
		}
	})

	t.Run("spoofed headers from an untrusted peer", func(t *testing.T) {
		got := declineFrom(t, db, router, crossing{token: "tok-spoof", peer: "198.51.100.9:5555", edge: "203.0.113.66", visitor: "203.0.113.66"})
		if got != "198.51.100.9" {
			t.Fatalf("audit ip = %q, want the peer address", got)
		}
	})
}

// productionChain mounts the decision routes behind httpx.NewRouter configured
// from TRUSTED_PROXIES=private,cloudflare, as production runs.
func productionChain(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()
	t.Setenv("TRUSTED_PROXIES", "private,cloudflare")
	trusted, err := troncenv.TrustedProxies()
	if err != nil {
		t.Fatalf("TrustedProxies() = %v", err)
	}
	cdnProxies, cdnHeader := troncenv.CDN()

	smtpSvc := smtp.NewService(db)
	webhookSvc := webhooks.NewService(db)
	docSvc := documents.NewService(db, smtpSvc, webhookSvc, "http://localhost:5173", t.TempDir())
	service := NewService(db, docSvc, webhookSvc, smtpSvc, nil, "http://localhost:5173")

	router := httpx.NewRouter(httpx.Config{TrustedProxies: trusted, CDNProxies: cdnProxies, CDNHeader: cdnHeader})
	router.Route("/api", func(api chi.Router) { RegisterSigningDecisionRoutes(api, service) })
	return router
}

// crossing is one request as it reaches the API: the connecting peer, the edge
// Traefik wrote into X-Forwarded-For and X-Real-Ip, and the Cf-Connecting-Ip value.
type crossing struct {
	token, peer, edge, visitor string
}

// declineFrom seeds a pending signer, declines through router as described by
// c, and returns the IP the audit row recorded.
func declineFrom(t *testing.T, db *gorm.DB, router http.Handler, c crossing) string {
	t.Helper()
	doc := schemas.Document{Name: "Contract", Status: "pending", FileName: "contract.pdf", OwnerID: 1}
	if err := db.Create(&doc).Error; err != nil {
		t.Fatalf("create document: %v", err)
	}
	signer := schemas.Signer{DocumentID: doc.ID, Name: "Alice", Email: "alice@example.com", Role: "signer", Status: "pending", Token: c.token}
	if err := db.Create(&signer).Error; err != nil {
		t.Fatalf("create signer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/sign/"+c.token+"/decline", nil)
	request.RemoteAddr = c.peer
	request.Header.Set("X-Forwarded-For", c.edge)
	request.Header.Set("X-Real-Ip", c.edge)
	request.Header.Set("Cf-Connecting-Ip", c.visitor)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("decline status = %d, body %s", recorder.Code, recorder.Body.String())
	}

	var after schemas.Signer
	if err := db.Where("id = ?", signer.ID).First(&after).Error; err != nil {
		t.Fatalf("reload signer: %v", err)
	}
	return after.IPAddress
}

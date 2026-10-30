package webhooks

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/schemas"
)

func decode(t *testing.T, p EventPayload) map[string]any {
	t.Helper()
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return out["document"].(map[string]any)
}

func TestCompletedEventCarriesTheLink(t *testing.T) {
	doc := &schemas.Document{
		ID: 12, Name: "Contract", Status: "completed", ExternalRef: "ardoise:contract:42",
		NuageSpaceFacileID: "fac_0123456789abcdef01234567", NuageLinkedTo: "ardoise:contract:42:signed",
	}
	got := decode(t, BuildDocumentEvent(EventDocumentCompleted, doc, "http://plume.test"))
	nuage := got["nuage"].(map[string]any)
	if got["id"].(float64) != 12 || got["status"] != "completed" || got["external_ref"] != "ardoise:contract:42" {
		t.Fatalf("document = %v", got)
	}
	if nuage["space_facile_id"] != "fac_0123456789abcdef01234567" || nuage["linked_to"] != "ardoise:contract:42:signed" || nuage["published"] != false {
		t.Fatalf("nuage = %v", nuage)
	}
}

func TestPlainDocumentEventsAreUnchanged(t *testing.T) {
	got := decode(t, BuildDocumentEvent(EventDocumentCompleted, &schemas.Document{ID: 1, Status: "completed"}, ""))
	if _, ok := got["external_ref"]; ok {
		t.Error("external_ref leaked into an unlinked event")
	}
	if _, ok := got["nuage"]; ok {
		t.Error("nuage leaked into an unlinked event")
	}
}

func TestArchivedEventReportsThePublishedTarget(t *testing.T) {
	now := time.Now()
	doc := &schemas.Document{ID: 5, Status: "completed", ExternalRef: "r", NuagePublishedAt: &now}
	payload := BuildArchivedEvent(doc, "", "fac_0123456789abcdef01234567", "plume:document:5")
	got := decode(t, payload)
	nuage := got["nuage"].(map[string]any)
	if payload.EventType != "document.archived" || nuage["published"] != true || nuage["linked_to"] != "plume:document:5" {
		t.Fatalf("event %s nuage %v", payload.EventType, nuage)
	}
}

func TestSignatureIsHMACSHA256OfTheBody(t *testing.T) {
	const want = "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
	if got := sign("key", []byte("The quick brown fox jumps over the lazy dog")); got != want {
		t.Fatalf("sign = %q, want %q", got, want)
	}
}

package signers

import (
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
)

func TestNuageJobTargetsTheOwnersFolderAndKeepsTheRequestedKey(t *testing.T) {
	cases := []struct {
		name       string
		doc        schemas.Document
		wantLinked string
	}{
		{"plain document", schemas.Document{ID: 3, OwnerID: 8, Name: "n"}, ""},
		{"requested key", schemas.Document{ID: 3, OwnerID: 8, Name: "n", NuageLinkedTo: "ardoise:1"}, "ardoise:1"},
		{"a requested space no longer routes anything", schemas.Document{ID: 3, OwnerID: 8, Name: "n", NuageSpaceFacileID: "fac_bbbbbbbbbbbbbbbbbbbbbbbb"}, ""},
	}
	for _, c := range cases {
		job := nuageJob(&c.doc)
		if job.DocumentID != 3 || job.OwnerID != 8 || job.Title != "n" || job.LinkedTo != c.wantLinked {
			t.Errorf("%s: got %+v", c.name, job)
		}
	}
}

package signers

import (
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
)

func TestNuageJobSelectsTheTarget(t *testing.T) {
	space := &schemas.Space{Name: "Studio", FacileID: "fac_aaaaaaaaaaaaaaaaaaaaaaaa"}
	explicit := "fac_bbbbbbbbbbbbbbbbbbbbbbbb"
	cases := []struct {
		name         string
		doc          schemas.Document
		space        *schemas.Space
		wantSpace    string
		wantSpaceNam string
		wantLinked   string
	}{
		{"plain document keeps today's behaviour", schemas.Document{ID: 3}, nil, "", "", ""},
		{"plume space is used", schemas.Document{ID: 3}, space, space.FacileID, "Studio", ""},
		{"explicit space wins over the plume space", schemas.Document{ID: 3, NuageSpaceFacileID: explicit}, space, explicit, "", ""},
		{"explicit key alone keeps the plume space", schemas.Document{ID: 3, NuageLinkedTo: "ardoise:1"}, space, space.FacileID, "Studio", "ardoise:1"},
		{"both set", schemas.Document{ID: 3, NuageSpaceFacileID: explicit, NuageLinkedTo: "ardoise:1"}, nil, explicit, "", "ardoise:1"},
	}
	for _, c := range cases {
		job := nuageJob(&c.doc, c.space)
		if job.SpaceFacileID != c.wantSpace || job.SpaceName != c.wantSpaceNam || job.LinkedTo != c.wantLinked || job.DocumentID != 3 {
			t.Errorf("%s: got %+v", c.name, job)
		}
	}
}

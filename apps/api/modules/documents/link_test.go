package documents

import (
	"strings"
	"testing"
)

func TestLinkValidate(t *testing.T) {
	good := "fac_0123456789abcdef01234567"
	cases := []struct {
		name string
		link Link
		ok   bool
	}{
		{"empty is fine", Link{}, true},
		{"all set", Link{ExternalRef: "ardoise:contract:42", NuageSpaceFacileID: good, NuageLinkedTo: "ardoise:contract:42:signed"}, true},
		{"200 characters", Link{ExternalRef: strings.Repeat("é", 200)}, true},
		{"201 characters", Link{ExternalRef: strings.Repeat("a", 201)}, false},
		{"control character", Link{ExternalRef: "a\nb"}, false},
		{"tab", Link{NuageLinkedTo: "a\tb"}, false},
		{"surrounding space", Link{ExternalRef: " ref"}, false},
		{"facile id uppercase", Link{NuageSpaceFacileID: "fac_0123456789ABCDEF01234567"}, false},
		{"facile id short", Link{NuageSpaceFacileID: "fac_abc"}, false},
		{"facile id wrong prefix", Link{NuageSpaceFacileID: "xxx_0123456789abcdef01234567"}, false},
		{"linked_to too long", Link{NuageLinkedTo: strings.Repeat("a", 201)}, false},
	}
	for _, c := range cases {
		err := c.link.Validate()
		if (err == nil) != c.ok {
			t.Errorf("%s: err = %v, want ok=%v", c.name, err, c.ok)
		}
	}
}

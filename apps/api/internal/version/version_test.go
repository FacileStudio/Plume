package version

import (
	"encoding/json"
	"os"
	"runtime/debug"
	"testing"
)

func TestFromBuildInfoPrefersTheModuleTag(t *testing.T) {
	bi := &debug.BuildInfo{}
	bi.Main.Version = "v1.4.2"
	if got := fromBuildInfo(bi); got != "v1.4.2" {
		t.Fatalf("fromBuildInfo = %q, want v1.4.2", got)
	}
}

func TestFromBuildInfoFallsBackToTheRevision(t *testing.T) {
	bi := &debug.BuildInfo{Settings: []debug.BuildSetting{
		{Key: "vcs.revision", Value: "0123456789abcdef0123"},
		{Key: "vcs.modified", Value: "true"},
	}}
	bi.Main.Version = "(devel)"
	if got := fromBuildInfo(bi); got != "0123456789ab+dirty" {
		t.Fatalf("fromBuildInfo = %q, want 0123456789ab+dirty", got)
	}
}

func TestFromBuildInfoWithNoStampSaysDev(t *testing.T) {
	bi := &debug.BuildInfo{}
	bi.Main.Version = "(devel)"
	if got := fromBuildInfo(bi); got != "dev" {
		t.Fatalf("fromBuildInfo = %q, want dev", got)
	}
}

func TestAStampOutranksTheToolchain(t *testing.T) {
	stamp = "0123456789ab"
	t.Cleanup(func() { stamp = "" })
	if got := String(); got != "0123456789ab" {
		t.Fatalf("String = %q, want the linked-in stamp", got)
	}
}

func TestTheLinkedStampMatchesPackageJSON(t *testing.T) {
	if stamp == "" {
		t.Skip("built without -X stamp; scripts/check.sh links it from package.json")
	}
	raw, err := os.ReadFile("../../../client/package.json")
	if err != nil {
		t.Fatal(err)
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		t.Fatal(err)
	}
	if want := "v" + pkg.Version; String() != want {
		t.Fatalf("String = %q, package.json says %q", String(), want)
	}
}

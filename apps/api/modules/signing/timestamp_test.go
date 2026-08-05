package signing

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestFormatEventTimestampConvertsToUTC(t *testing.T) {
	zone := time.FixedZone("TEST", 5*3600)
	local := time.Date(2026, time.August, 5, 2, 30, 0, 0, zone)

	got := formatEventTimestamp(local)
	want := "2026-08-04 21:30 UTC"

	if got != want {
		t.Fatalf("formatEventTimestamp(%s) = %q, want %q", local, got, want)
	}
}

func TestUTCLabelledLayoutsAreFedUTCTimes(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "service.go", nil, 0)
	if err != nil {
		t.Fatalf("parse service.go: %v", err)
	}

	found := 0
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Format" {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		layout, err := strconv.Unquote(lit.Value)
		if err != nil || !strings.Contains(layout, "UTC") {
			return true
		}

		found++
		inner, ok := sel.X.(*ast.CallExpr)
		if !ok {
			t.Errorf("%s: .Format(%q) on a time that was never converted with .UTC()", fset.Position(call.Pos()), layout)
			return true
		}
		innerSel, ok := inner.Fun.(*ast.SelectorExpr)
		if !ok || innerSel.Sel.Name != "UTC" {
			t.Errorf("%s: .Format(%q) on a time that was never converted with .UTC()", fset.Position(call.Pos()), layout)
		}
		return true
	})

	if found == 0 {
		t.Fatal("no UTC-labelled Format layouts found in service.go; guard is no longer wired to anything")
	}
}

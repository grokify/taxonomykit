package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const samplePSV = `lang | parent | id | slug | name
en | | all | all | All
en | all | law | law | Law
es | all | law | derecho | Derecho
en | law | patent | patent | Patent
en | all | business | business | Business
en | business | realestate | realestate | Real Estate
en | business | finance | finance | Finance
en | finance | mortgage | mortgage | Mortgage
en | realestate | mortgage | mortgage | Mortgage
`

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := newRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func writeSample(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestValidateAndConvert(t *testing.T) {
	psv := writeSample(t, "topics.psv", samplePSV)
	out, err := run(t, "validate", psv)
	if err != nil || !strings.HasPrefix(out, "ok: 7 nodes, languages en,es") {
		t.Fatalf("validate: %v %q", err, out)
	}

	jsonPath := filepath.Join(t.TempDir(), "topics.json")
	if _, err := run(t, "convert", psv, "--to", "json", "-o", jsonPath); err != nil {
		t.Fatal(err)
	}
	out, err = run(t, "convert", jsonPath, "--to", "psv")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "en | finance | mortgage | mortgage | Mortgage") || !strings.Contains(out, "en | realestate | mortgage | mortgage | Mortgage") {
		t.Fatalf("both parents survive the round trip:\n%s", out)
	}

	bad := writeSample(t, "bad.csv", "id,name,parent\na,A,\nb,B,zzz\n")
	out, err = run(t, "validate", bad)
	if err == nil || !strings.Contains(out, "parent.missing [b]") {
		t.Fatalf("validate should list problems and fail: %v %q", err, out)
	}
	if _, err := run(t, "convert", bad); err == nil {
		t.Fatal("convert must refuse an invalid document")
	}
}

func TestTreeAndLineage(t *testing.T) {
	psv := writeSample(t, "topics.psv", samplePSV)
	out, err := run(t, "tree", psv, "--lang", "es")
	if err != nil {
		t.Fatal(err)
	}
	want := "All (all) /all\n  Business (business) /business\n    Finance (finance) /finance\n      Mortgage (mortgage) /mortgage\n    Real Estate (realestate) /realestate\n  Derecho (law) /derecho\n    Patent (patent) /patent\n"
	if out != want {
		t.Fatalf("tree:\n got  %q\n want %q", out, want)
	}
	out, err = run(t, "tree", psv, "--root", "law", "--depth", "0")
	if err != nil || out != "Patent (patent) /patent\n" {
		t.Fatalf("subtree: %v %q", err, out)
	}
	out, err = run(t, "lineage", psv, "mortgage", "--all")
	if err != nil {
		t.Fatal(err)
	}
	if out != "All (all) > Business (business) > Finance (finance) > Mortgage (mortgage)\nAll (all) > Business (business) > Real Estate (realestate) > Mortgage (mortgage)\n" {
		t.Fatalf("lineage --all: %q", out)
	}
	if _, err := run(t, "lineage", psv, "nope"); err == nil {
		t.Fatal("unknown node")
	}
	if _, err := run(t, "tree", writeSample(t, "noext", samplePSV)); err == nil {
		t.Fatal("format inference must fail without an extension or --from")
	}
	if _, err := run(t, "tree", writeSample(t, "noext", samplePSV), "--from", "psv"); err != nil {
		t.Fatal(err)
	}
}

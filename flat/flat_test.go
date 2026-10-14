package flat

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/grokify/taxonomykit"
)

const legacyPSV = `# version | lang | parent | id | slug | name
lang | parent | id | slug | name
en | | all | all | All
en | all | business | business | Business
en | all | law | law | Law
en | business | realestate | realestate | Real Estate
en | business | finance | finance | Finance
en | finance | personal-finance | personal-finance | Personal Finance

en | realestate | mortgage | mortgage | Mortgage
en | personal-finance | mortgage | mortgage | Mortgage
es | all | law | derecho | Derecho
es | | all | todo | Todo
`

func TestReadPSVMergesRows(t *testing.T) {
	doc, err := Read(strings.NewReader(legacyPSV), PSV, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if doc.DefaultLang != "en" || doc.Version != taxonomykit.DocumentVersion {
		t.Fatalf("doc header: %+v", doc)
	}
	byID := map[string]taxonomykit.Node{}
	for _, n := range doc.Nodes {
		byID[n.ID] = n
	}
	if got := byID["mortgage"].Parents; !reflect.DeepEqual(got, []string{"realestate", "personal-finance"}) {
		t.Fatalf("repeated rows add parents in order: %v", got)
	}
	if got := byID["law"].Labels["es"]; got.Name != "Derecho" || got.Slug != "derecho" {
		t.Fatalf("second language row: %+v", got)
	}
	if got := byID["all"].Parents; len(got) != 0 {
		t.Fatalf("root keeps no parents: %v", got)
	}
	if doc.Nodes[0].ID != "all" || doc.Nodes[len(doc.Nodes)-1].ID != "mortgage" {
		t.Fatal("first-appearance order")
	}
	if _, err := taxonomykit.Build(doc.Nodes, taxonomykit.WithDefaultLang(doc.DefaultLang)); err != nil {
		t.Fatalf("legacy file should build: %v", err)
	}
}

func TestReadCSVWithAttrsAltOrder(t *testing.T) {
	in := "id,name,slug,parent,alt,order,color\n" +
		"r,Root,root,,,,red\n" +
		"a,\"A, the first\",a,r,alpha;first,2,\n" +
		"b,B,b,r,,1,blue\n"
	doc, err := Read(strings.NewReader(in), CSV, Options{DefaultLang: "en"})
	if err != nil {
		t.Fatal(err)
	}
	a := doc.Nodes[1]
	if a.Labels["en"].Name != "A, the first" || !reflect.DeepEqual(a.Labels["en"].Alt, []string{"alpha", "first"}) || *a.Order != 2 {
		t.Fatalf("csv row: %+v", a)
	}
	if doc.Nodes[0].Attrs["color"] != "red" || doc.Nodes[1].Attrs != nil {
		t.Fatalf("attrs: %+v %+v", doc.Nodes[0].Attrs, doc.Nodes[1].Attrs)
	}
	tr := taxonomykit.MustBuild(doc.Nodes)
	if got := tr.Children("r"); got[0].ID != "b" || got[1].ID != "a" {
		t.Fatal("order column drives sibling order")
	}
}

func TestReadHeaderless(t *testing.T) {
	in := "\tr\tRoot\nr\ta\tA\n"
	cols := &Columns{Parent: 0, ID: 1, Name: 2, Lang: -1, Slug: -1, Alt: -1, Order: -1}
	doc, err := Read(strings.NewReader(in), TSV, Options{Columns: cols, DefaultLang: "fr"})
	if err != nil {
		t.Fatal(err)
	}
	if doc.DefaultLang != "fr" || doc.Nodes[1].Labels["fr"].Name != "A" || doc.Nodes[1].Parents[0] != "r" {
		t.Fatalf("headerless: %+v", doc)
	}
}

func TestReadErrors(t *testing.T) {
	if _, err := Read(strings.NewReader("id,name\n,x\n"), CSV, Options{}); err == nil {
		t.Fatal("empty id")
	}
	if _, err := Read(strings.NewReader("id,name,order\na,A,two\n"), CSV, Options{}); err == nil {
		t.Fatal("bad order")
	}
	if _, err := Read(strings.NewReader("name\nx\n"), CSV, Options{}); err == nil {
		t.Fatal("missing id column")
	}
	if _, err := Read(strings.NewReader("id,,name\n"), CSV, Options{}); err == nil {
		t.Fatal("empty header cell")
	}
	if _, err := ParseFormat("xlsx"); err == nil {
		t.Fatal("unknown format")
	}
	if f, err := ParseFormat(".PSV"); err != nil || f != PSV || f.String() != "psv" {
		t.Fatal("parse format")
	}
}

func TestWriteRoundTrip(t *testing.T) {
	doc, err := Read(strings.NewReader(legacyPSV), PSV, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []Format{CSV, TSV, PSV} {
		var buf bytes.Buffer
		if err := Write(&buf, doc, f); err != nil {
			t.Fatal(err)
		}
		first := buf.String()
		back, err := Read(strings.NewReader(first), f, Options{DefaultLang: doc.DefaultLang})
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if !reflect.DeepEqual(back, doc) {
			t.Fatalf("%s round trip changed the document:\n%s", f, first)
		}
		var again bytes.Buffer
		if err := Write(&again, back, f); err != nil {
			t.Fatal(err)
		}
		if again.String() != first {
			t.Fatalf("%s: second write differs", f)
		}
	}
}

func TestWriteRejectsDelimiterInCell(t *testing.T) {
	doc := &taxonomykit.Document{Version: "1", DefaultLang: "en", Nodes: []taxonomykit.Node{
		{ID: "a", Labels: taxonomykit.Labels{"en": {Name: "A | B"}}},
	}}
	if err := Write(&bytes.Buffer{}, doc, PSV); err == nil {
		t.Fatal("pipe in a PSV cell must be rejected")
	}
	if err := Write(&bytes.Buffer{}, doc, CSV); err != nil {
		t.Fatal("csv quotes it")
	}
}

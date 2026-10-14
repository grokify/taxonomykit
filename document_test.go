package taxonomykit

import (
	"reflect"
	"strings"
	"testing"
)

func TestDocumentJSONRoundTrip(t *testing.T) {
	doc := NewDocument("EN", fixture())
	if doc.DefaultLang != "en" {
		t.Fatal("NewDocument normalizes the default language")
	}
	data, err := doc.JSON()
	if err != nil {
		t.Fatal(err)
	}
	back, err := ParseDocument(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(back, doc) {
		t.Fatal("round trip changed the document")
	}
	again, err := back.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(data) {
		t.Fatal("encoding is not deterministic")
	}
	if ps := doc.Validate(); ps != nil {
		t.Fatal(ps)
	}
	tr, err := doc.Build()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tr.Document(), doc) {
		t.Fatal("Tree.Document should reproduce the input document")
	}
}

func TestParseDocumentStrict(t *testing.T) {
	if _, err := ParseDocument([]byte(`{"version":"1","defaultLang":"en","nodes":[],"extra":1}`)); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown fields must be rejected: %v", err)
	}
	if _, err := ParseDocument([]byte(`{"version":"2","defaultLang":"en","nodes":[]}`)); err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("unsupported version must be rejected: %v", err)
	}
	if _, err := ParseDocument([]byte(`{`)); err == nil {
		t.Fatal("malformed JSON")
	}
}

package taxonomykit

import (
	"reflect"
	"testing"
)

func TestNormalizeLangAndFallback(t *testing.T) {
	if NormalizeLang(" EN_us ") != "en-us" {
		t.Fatal("normalize")
	}
	if got := FallbackChain("pt-BR-x", "en"); !reflect.DeepEqual(got, []string{"pt-br-x", "pt-br", "pt", "en"}) {
		t.Fatalf("chain: %v", got)
	}
	if got := FallbackChain("en-us", "en"); !reflect.DeepEqual(got, []string{"en-us", "en"}) {
		t.Fatalf("chain ending at default: %v", got)
	}
	if got := FallbackChain("", "en"); !reflect.DeepEqual(got, []string{"en"}) {
		t.Fatalf("empty: %v", got)
	}
}

func TestLabelFallback(t *testing.T) {
	tr := MustBuild(fixture())
	l, used, ok := tr.Label("law", "es")
	if !ok || used != "es" || l.Name != "Derecho" {
		t.Fatalf("exact: %v %s %v", l, used, ok)
	}
	l, used, ok = tr.Label("patent", "es")
	if !ok || used != "en" || l.Name != "Patent" {
		t.Fatalf("fallback to default: %v %s %v", l, used, ok)
	}
	l, used, ok = tr.Label("law", "es-MX")
	if !ok || used != "es" || l.Name != "Derecho" {
		t.Fatalf("region → language: %v %s %v", l, used, ok)
	}
	if _, _, ok := tr.Label("nope", "en"); ok {
		t.Fatal("unknown id")
	}
	if tr.Name("law", "fr") != "Law" || tr.Slug("law", "es") != "derecho" || tr.Slug("patent", "es") != "patent" {
		t.Fatal("Name/Slug helpers")
	}
}

func TestBreadcrumb(t *testing.T) {
	tr := MustBuild(fixture())
	got := tr.Breadcrumb("patent", "es")
	want := []Crumb{
		{ID: "all", Name: "Todo", Slug: "todo", Lang: "es"},
		{ID: "law", Name: "Derecho", Slug: "derecho", Lang: "es"},
		{ID: "patent", Name: "Patent", Slug: "patent", Lang: "en"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("breadcrumb:\n got  %+v\n want %+v", got, want)
	}
	if tr.Breadcrumb("nope", "en") != nil {
		t.Fatal("unknown id")
	}
}

func TestBySlug(t *testing.T) {
	tr := MustBuild(fixture())
	if n, ok := tr.BySlug("es", "derecho"); !ok || n.ID != "law" {
		t.Fatal("explicit es slug")
	}
	if n, ok := tr.BySlug("es", "patent"); !ok || n.ID != "patent" {
		t.Fatal("fallback slug resolves in es")
	}
	if _, ok := tr.BySlug("es", "law"); ok {
		t.Fatal("the en slug of a node with an explicit es slug must not resolve in es")
	}
	if n, ok := tr.BySlug("es-MX", "derecho"); !ok || n.ID != "law" {
		t.Fatal("unknown region resolves through the chain")
	}
	if n, ok := tr.BySlug("fr", "patent"); !ok || n.ID != "patent" {
		t.Fatal("unknown language resolves through the default")
	}
	if _, ok := tr.BySlug("en", "nope"); ok {
		t.Fatal("unknown slug")
	}
	if got := tr.Langs(); !reflect.DeepEqual(got, []string{"en", "es"}) {
		t.Fatalf("langs: %v", got)
	}
	if tr.DefaultLang() != "en" {
		t.Fatal("default lang")
	}
}

func TestDefaultLangOption(t *testing.T) {
	nodes := []Node{{ID: "a", Labels: Labels{"ES": {Name: "Hola", Slug: "hola"}}}}
	if _, err := Build(nodes); err == nil {
		t.Fatal("no en label should fail under the default")
	}
	tr, err := Build(nodes, WithDefaultLang("es"))
	if err != nil {
		t.Fatal(err)
	}
	if tr.Name("a", "en") != "Hola" {
		t.Fatal("fallback to es default")
	}
}

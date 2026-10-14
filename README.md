# TaxonomyKit

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
![coverage](https://img.shields.io/badge/coverage-93.1%25-brightgreen)
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/grokify/taxonomykit/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/grokify/taxonomykit/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/grokify/taxonomykit/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/grokify/taxonomykit/actions/workflows/go-lint.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/taxonomykit
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/taxonomykit
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/taxonomykit/blob/main/LICENSE

Hierarchical taxonomy toolkit for Go: load, validate, navigate, localize, and convert category trees — structure only, you bring the content.

## Status

v0.1 in progress: the core tree, codecs, schema and CLI are implemented; SKOS interchange and document operations follow in v0.2. The design is specified in [`docs/specs/`](docs/specs/) (PRD, TRD, PLAN, ROADMAP).

## Why

Most applications that classify things — content sites, product catalogs, knowledge bases, tagging systems — end up with the same ad-hoc code: a flat list of categories with parent pointers, a function that walks up to build breadcrumbs, another that lists children for a menu, a lookup by slug, and a localized display name bolted on later. TaxonomyKit is that code once, tested, with the cases ad-hoc versions get wrong: multiple parents, locale fallback, deterministic ordering, cycle and orphan detection.

The model is deliberately close to [SKOS](https://www.w3.org/TR/skos-reference/) (concepts with `broader`/`narrower` relations and per-language preferred and alternative labels), so taxonomies can be exchanged with the rest of the world without a redesign.

## Usage

```go
import "github.com/grokify/taxonomykit"

tree, err := taxonomykit.Build([]taxonomykit.Node{
    {ID: "all", Labels: taxonomykit.Labels{"en": {Name: "All", Slug: "all"}}},
    {ID: "law", Parents: []string{"all"}, Labels: taxonomykit.Labels{
        "en": {Name: "Law", Slug: "law"},
        "es": {Name: "Derecho", Slug: "derecho"},
    }},
    {ID: "patent", Parents: []string{"law"}, Labels: taxonomykit.Labels{"en": {Name: "Patent", Slug: "patent"}}},
}, taxonomykit.WithDefaultLang("en"))
if err != nil {
    return err // cycles, orphans, duplicate ids or slugs, missing default labels
}

tree.Lineage("patent")          // all → law → patent (through primary parents)
tree.Lineages("mortgage")       // every root path of a multi-parent node
tree.Breadcrumb("patent", "es") // Derecho(/derecho, es) › Patent(/patent, en) — each crumb says which language it fell back to
tree.Children("all")            // ordered: explicit Order, then default-language name, then id
tree.Siblings("patent")
tree.IsLeaf("patent")           // true
tree.BySlug("es", "derecho")    // law
```

The same example runs as `ExampleBuild` in `example_test.go`.

### Files

```go
doc, err := flat.Read(f, flat.PSV, flat.Options{})       // or CSV / TSV; header row or flat.Columns
tree, err := doc.Build()                                 // uses the document's default language
data, err := doc.JSON()                                  // deterministic JSON; schema in package schema
```

Flat files have one row per node and language. A node listed twice with different parents becomes a multi-parent node rather than silently keeping the last parent:

```text
lang | parent           | id       | slug     | name
en   |                  | all      | all      | All
en   | all              | law      | law      | Law
es   | all              | law      | derecho  | Derecho
en   | finance          | mortgage | mortgage | Mortgage
en   | realestate       | mortgage | mortgage | Mortgage
```

### CLI

```sh
go install github.com/grokify/taxonomykit/cmd/taxonomykit@latest
taxonomykit validate topics.psv              # every problem, exit 1 if any
taxonomykit convert topics.psv --to json     # csv | tsv | psv | json
taxonomykit tree topics.json --lang es       # indented listing with slugs
taxonomykit lineage topics.psv mortgage --all
```

## License

MIT — see [LICENSE](LICENSE).

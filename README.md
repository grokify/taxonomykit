# TaxonomyKit

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
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

Pre-release. The design is specified in [`docs/specs/`](docs/specs/) (PRD, TRD, PLAN, ROADMAP); implementation follows the roadmap phases.

## Why

Most applications that classify things — content sites, product catalogs, knowledge bases, tagging systems — end up with the same ad-hoc code: a flat list of categories with parent pointers, a function that walks up to build breadcrumbs, another that lists children for a menu, a lookup by slug, and a localized display name bolted on later. TaxonomyKit is that code once, tested, with the cases ad-hoc versions get wrong: multiple parents, locale fallback, deterministic ordering, cycle and orphan detection.

The model is deliberately close to [SKOS](https://www.w3.org/TR/skos-reference/) (concepts with `broader`/`narrower` relations and per-language preferred and alternative labels), so taxonomies can be exchanged with the rest of the world without a redesign.

## Intended API (v0.1 target)

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

tree.Lineage("patent")          // [all law patent]
tree.Breadcrumb("patent", "es") // [{Derecho derecho law} {Patent patent patent}] — falls back to "en"
tree.Children("all")            // ordered deterministically
tree.Siblings("patent")
tree.IsLeaf("patent")           // true
tree.BySlug("es", "derecho")    // law
```

A flat CSV/TSV/PSV codec, a JSON document format with a generated JSON Schema, and a `taxonomykit` CLI (`validate`, `convert`, `tree`) are part of v0.1; SKOS JSON-LD import/export follows.

## License

MIT — see [LICENSE](LICENSE).

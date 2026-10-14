# TaxonomyKit

**Hierarchical taxonomy toolkit for Go.** Load a category tree from the file
you already keep it in, validate it, navigate it, localize it, convert it —
structure only, you bring the content.

Most applications that classify things end up rewriting the same tree code:
a flat list of categories with parent pointers, a function that walks up to
build breadcrumbs, another that lists children for a menu, a lookup by slug,
and a second language bolted on later. TaxonomyKit is that code once, with
the cases ad-hoc versions get wrong handled: nodes with several parents,
locale fallback, deterministic ordering, cycle and orphan detection.

!!! note "Structure, not content"
    TaxonomyKit never defines what your categories are. It takes nodes from
    a CSV, PSV or JSON file — or a Go slice — and gives you a validated tree
    to query. The model is deliberately close to
    [SKOS](https://www.w3.org/TR/skos-reference/), so taxonomies can be
    exchanged with the rest of the world without a redesign.

## A first look

```go
tree, err := taxonomykit.Build([]taxonomykit.Node{
    {ID: "all", Labels: taxonomykit.Labels{"en": {Name: "All", Slug: "all"}}},
    {ID: "law", Parents: []string{"all"}, Labels: taxonomykit.Labels{
        "en": {Name: "Law", Slug: "law"},
        "es": {Name: "Derecho", Slug: "derecho"},
    }},
    {ID: "patent", Parents: []string{"law"}, Labels: taxonomykit.Labels{"en": {Name: "Patent", Slug: "patent"}}},
})
if err != nil {
    return err
}
tree.Lineage("patent")          // all › law › patent
tree.Breadcrumb("patent", "es") // Derecho (/derecho, es) › Patent (/patent, en)
tree.BySlug("es", "derecho")    // law
```

Or from the command line:

```bash
taxonomykit validate topics.psv
taxonomykit tree topics.psv --lang es
```

## Where to go next

- [Installation](guides/installation.md) and [Quick Start](guides/quickstart.md)
- [Concepts](guides/concepts.md) — nodes, primary parents, ordering, documents
- [File Formats](guides/file-formats.md) — CSV/TSV/PSV rows and the JSON document
- [Navigating a Tree](guides/navigation.md) and [Localization](guides/localization.md)
- [Validation](guides/validation.md) — every problem code and what to do about it
- [CLI Reference](guides/cli-reference.md)

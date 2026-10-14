# TaxonomyKit — Technical Requirements Document (TRD)

**Status:** Draft
**Date:** 2026-10-14
**Repo:** github.com/grokify/taxonomykit

## 1. Architecture overview

A small core package that builds an immutable, validated, indexed tree from a slice of nodes, surrounded by codecs that only produce or consume that slice.

```text
                    taxonomykit (core)
   ┌───────────────────────────────────────────────────┐
   │ Node · Label · Labels · Document                   │
   │ Build(nodes, opts) → *Tree  (validate + index)     │
   │ Tree: Node Roots Parent Parents Children Siblings  │
   │       Lineage Lineages Descendants Leaves IsLeaf   │
   │       Depth LCA Walk Breadcrumb BySlug Label        │
   │ Validate(nodes) → []Problem                        │
   └───────────────┬───────────────────┬───────────────┘
                   │                   │
        ┌──────────▼────────┐  ┌───────▼─────────┐   ┌──────────────┐
        │ flat/  CSV TSV PSV│  │ schema/ JSON     │   │ skos/ (v0.2) │
        │ read · write      │  │ Schema (go:embed)│   │ JSON-LD      │
        └──────────┬────────┘  └───────┬─────────┘   └──────┬───────┘
                   └───────────────────┼────────────────────┘
                              cmd/taxonomykit (validate · convert · tree)
```

Design rules:

- **Core imports only the standard library.** `cobra` in `cmd/` only.
- **`Tree` is immutable after `Build`.** Accessors return fresh slices (never internal storage); the tree is safe for concurrent reads.
- **Everything is ordered.** No map iteration reaches output. The ordering rule (§4) is applied once at build and cached.
- **Validation is total.** `Build` returns the first error; `Validate` returns every problem with node ids, so a CLI can list them all.
- **Errors are returned, never logged.** Programmer misuse (nil tree) panics.

## 2. Module and dependencies

- **Module:** `github.com/grokify/taxonomykit`, Go 1.26+
- **Library dependencies:** none
- **CLI:** `github.com/spf13/cobra`
- **Schema generation:** `github.com/invopop/jsonschema` in a `//go:build ignore` generator (`schema/gen/main.go`) with a `tools.go` guard; output linted with `schemakit lint --property-case camelCase`
- Verify versions at adoption time per convention.

## 3. Package layout

```text
taxonomykit/
├── node.go          # Node, Label, Labels, Document
├── build.go         # Build, Options (WithDefaultLang, WithSlugScope, WithOrderBy)
├── tree.go          # Tree navigation
├── locale.go        # tag normalization, fallback chain, Breadcrumb, BySlug
├── validate.go      # Validate, Problem, error types
├── walk.go          # Walk, DFS/BFS iterators (iter.Seq2)
├── flat/            # tabular codec: ReadCSV/TSV/PSV, Write*, Columns mapping, comment lines
├── schema/          # document.schema.json (generated), schema.go (//go:embed), gen/
├── skos/            # v0.2: JSON-LD import/export
├── cmd/taxonomykit/ # CLI
├── testdata/        # synthetic fixtures incl. every malformed case
└── docs/specs/
```

## 4. Core types and rules

```go
type Node struct {
    ID      string            `json:"id"`
    Parents []string          `json:"parents,omitempty"` // Parents[0] is the primary parent
    Labels  Labels            `json:"labels"`            // keyed by BCP-47 tag, lowercase ("en", "es", "pt-br")
    Order   *int              `json:"order,omitempty"`   // explicit sibling order; nil = sort by name
    Attrs   map[string]string `json:"attrs,omitempty"`   // opaque
}
type Label struct {
    Name string   `json:"name"`
    Slug string   `json:"slug,omitempty"` // empty means "no URL in this language"; never derived from Name
    Alt  []string `json:"alt,omitempty"`
}
type Labels map[string]Label

type Document struct {
    Version     string `json:"version"`     // "1"
    DefaultLang string `json:"defaultLang"`
    Nodes       []Node `json:"nodes"`
}
```

Rules enforced by `Build`/`Validate`:

| Rule | Problem code |
|------|--------------|
| `ID` non-empty, unique, matches `^[a-z0-9][a-z0-9._-]*$` | `id.invalid`, `id.duplicate` |
| Every parent id exists; a node is not its own parent | `parent.missing`, `parent.self` |
| No cycles through any parent edge | `cycle` |
| At least one root; roots have no parents | `root.none` |
| Every node has a label in the default language with a non-empty `Name` | `label.default.missing` |
| Per language, `Slug` unique across the tree (default scope) or among siblings (`WithSlugScope(SlugScopeSiblings)`) | `slug.duplicate` |
| Language tags normalize to lowercase BCP-47; unknown structure is rejected | `lang.invalid` |
| Duplicate parent ids on one node | `parent.duplicate` |

**Ordering rule (children, siblings, roots, descendants within a level):** nodes with `Order` set come first ascending by `Order`; then the rest ascending by default-language `Name` (case-insensitive, then case-sensitive), then by `ID`. `WithOrderBy(OrderByID)` switches the name step to id-only.

**Locale fallback for `Label(id, lang)`:** exact tag → language subtag (`pt-br` → `pt`) → default language. `Breadcrumb` applies it per crumb and reports in each `Crumb` which language was used (`Lang` field), so a site can decide whether to show a fallback indicator.

**Lineage with multiple parents:** `Lineage` follows `Parents[0]` at every step; `Lineages` enumerates all root paths (deduplicated, ordered by the ordering rule applied to each path's second element). `Depth` is the length of the primary lineage minus one. `LCA` operates on primary lineages.

**Descendants / Leaves / Walk:** DFS pre-order through `Children`; a node reachable by two paths is visited once (first occurrence). `Walk` and `All`/`Descend` iterators use `iter.Seq2[*Node, int]` with depth.

## 5. Flat codec (`flat/`)

- Input: CSV, TSV, or PSV (`|`-separated, surrounding whitespace trimmed) with a header row, or a `Columns` mapping for headerless files. Lines starting with `#` and blank lines are skipped. Columns: `lang`, `parent` (empty = root), `id`, `slug` (optional), `name`, `alt` (optional, `;`-separated), `order` (optional); unknown columns go to `Attrs`.
- One row per (node, language). The first row for a node sets `Parents`; later rows may repeat or omit the parent; a *different* parent on a later row **adds** a parent (polyhierarchy by repetition) — this is how legacy single-parent files that listed a node twice become multi-parent documents rather than silently last-wins. `Parents[0]` is the first parent seen.
- Output: `Write*` emits one row per (node, language) in tree order, deterministic.

## 6. JSON document and schema (`schema/`)

`Document` round-trips through `encoding/json`; `schema/document.schema.json` is generated from the Go types (Go-first), embedded via `//go:embed`, and exposed as `schema.Document() []byte` plus `schema.Validate(doc)` using the standard library's JSON decoding followed by `Validate(nodes)` (no third-party schema validator in the library).

## 7. CLI (`cmd/taxonomykit`)

```text
taxonomykit validate <file>                 # all problems, exit 1 if any
taxonomykit convert <file> --to csv|tsv|psv|json [--columns …] [-o out]
taxonomykit tree <file> [--lang xx] [--root id] [--depth n]   # indented listing with slugs
taxonomykit lineage <file> <id> [--lang xx] [--all]
```

Input format is inferred from the extension, overridable with `--from`.

## 8. SKOS (`skos/`, v0.2)

- Export: one `skos:ConceptScheme`, one `skos:Concept` per node with `skos:prefLabel` (language-tagged) per label, `skos:altLabel` for `Alt`, `skos:broader` for every parent, `skos:narrower` inverse, `skos:topConceptOf` for roots; the primary parent recorded as `tk:primaryBroader` in a kit namespace. JSON-LD with a fixed `@context`.
- Import: the same JSON-LD subset; other serializations are out of scope (no Turtle/RDF-XML parser).

## 9. Testing

- Table-driven tests per rule in §4 with a fixture per problem code in `testdata/invalid/`.
- Property tests (`testing/quick`-style, seeded): lineage/children inverse relations, leaves ∪ internal = all, deterministic output across shuffled input order.
- Golden files for flat and JSON output; CLI tests via `os/exec` on `testdata/`.
- Benchmarks: `Build` and `Breadcrumb` at 1k and 10k nodes; target < 50 ms build at 10k.
- `golangci-lint` clean; coverage badge; `CGO_ENABLED=0` build.

## 10. Publication hygiene

Public from the first push: synthetic fixtures only, generic commit messages, no consumer names or data.

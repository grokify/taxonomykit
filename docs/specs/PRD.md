# TaxonomyKit — Product Requirements Document (PRD)

**Status:** Draft
**Date:** 2026-10-14
**Repo:** github.com/grokify/taxonomykit

## 1. Summary

TaxonomyKit is a Go toolkit for hierarchical taxonomies — category trees, topic trees, product hierarchies, subject classifications. It owns the **structure** (nodes, parent relations, ordering, per-language labels and slugs) and the operations applications need on it (lineage, children, siblings, leaves, descendants, breadcrumbs, slug lookup, validation, conversion); the **content** stays with the application.

It follows the `-kit` convention of this organization: a toolkit for working with a kind of artifact the caller already has, in whatever file they keep it in.

## 2. Problem

Every classifying application rewrites the same tree code and gets the same things wrong:

- **Single-parent assumption.** Real taxonomies have nodes that belong in two places (a "Mortgage" topic under both Real Estate and Personal Finance). Ad-hoc loaders keyed by id silently keep the last parent they saw and lose the other link.
- **Localization bolted on.** Display names in a second language are added as a parallel table, slugs are either not localized or collide, and there is no fallback rule when a label is missing.
- **Non-deterministic order.** Children come out in map order; menus and sitemaps change between builds.
- **No validation.** Cycles, orphans, duplicate ids, and duplicate slugs are discovered in production.
- **Three ids per node.** Short codes for humans, slugs for URLs, and sanitized ids for a search index accrete over time with no rule for which is canonical.
- **No interchange.** The tree lives in a bespoke CSV; there is no path to or from SKOS, the W3C standard for exactly this.

## 3. Target users

| User | Need |
|------|------|
| **Content sites** (glossaries, docs, catalogs) | Breadcrumbs, topic menus, "also listed in", letter/topic index pages, localized URLs |
| **Product and data teams** | A validated, versioned taxonomy file with a schema, convertible between CSV, JSON, and SKOS |
| **Library authors** | A small dependency-free tree they can embed for tagging, faceting, or routing |

The first consumer is a private content site with a ~150-node, two-language topic tree that previously handled this in PHP and Ruby.

## 4. Goals (v0.1)

- **G1 — Correct tree semantics.** Polyhierarchy (multiple parents) with an explicit primary parent; lineage via primary parents, all lineages on request; cycle, orphan, and duplicate detection at build time.
- **G2 — Localization built in.** Per-language name, slug, and alternative labels on every node; a defined fallback chain (exact tag → language subtag → default language); per-language slug uniqueness; slug lookup per language.
- **G3 — Determinism.** Every slice the tree returns is in a defined order (explicit order, then name in the default language, then id). Same input ⇒ same output, byte for byte, in every codec.
- **G4 — One canonical id.** Nodes are identified by a single stable, locale-independent id; slugs and codes are labels, not identities.
- **G5 — Interchange.** A flat tabular codec (CSV/TSV/PSV with a language column), a JSON document format with a Go-generated JSON Schema, and a CLI to validate, convert, and print trees. SKOS JSON-LD import/export in v0.2.
- **G6 — Engine-only, dependency-free.** Standard library only in library packages (`cobra` in `cmd/` only); no I/O in the core beyond `io.Reader`/`io.Writer`.

## 5. Non-goals

- **Not a CMS or admin UI.** No editing workflows, no persistence layer; the taxonomy file is the source of truth and lives in the caller's repo or database.
- **Not an ontology engine.** No inference, no arbitrary relation types beyond broader/narrower and related; no OWL.
- **Not a tagging or classification service.** Assigning items to nodes, counting items per node, and faceted search belong to the application; the kit only provides the tree those features walk.
- **Not a faceted taxonomy (multiple independent trees in one document) in v0.1.** One scheme per document; multiple schemes are a v0.2 question.

## 6. Core concepts

- **Node** — the unit of classification: `ID` (canonical, stable, locale-independent), `Parents` (zero for a root, one or more otherwise; first is the primary), `Labels` by language tag (`Name`, URL `Slug`, optional `Alt` names), optional explicit `Order`, and `Attrs` (opaque string map the kit never interprets).
- **Tree** — the validated, immutable, indexed result of `Build`. All navigation operates on it.
- **Lineage** — the path root → node through primary parents (breadcrumb path). `Lineages` returns every path for multi-parent nodes.
- **Breadcrumb** — a lineage resolved for a language: name and slug per crumb with fallback applied.
- **Document** — the serializable form: `{version, defaultLang, nodes[]}`; the flat codec is a projection of the same data.
- **Scheme** — SKOS term for one taxonomy; a document is one scheme.

## 7. Key user flows

```go
// Load, validate, navigate
doc, err := flat.ReadPSV(r, flat.Columns{Lang: "lang", Parent: "parent", ID: "id", Slug: "slug", Name: "name"})
tree, err := taxonomykit.Build(doc.Nodes, taxonomykit.WithDefaultLang(doc.DefaultLang))

for _, c := range tree.Breadcrumb("mortgage", "en") { /* render */ }
tree.Lineages("mortgage")        // [[all business realestate mortgage] [all business finance personal-finance mortgage]]
tree.Children("technology")      // ordered
tree.Leaves("science")           // every leaf under science
tree.BySlug("es", "derecho")     // law

// Convert and validate from the command line
taxonomykit validate topics.psv
taxonomykit convert topics.psv --to json > topics.json
taxonomykit tree topics.json --lang es
```

## 8. Success metrics

| Dimension | Target |
|-----------|--------|
| Correctness | Property tests: every node appears in exactly one `Lineage`; `Children`/`Parent` are inverses; `Leaves` ∪ internal = all; validation rejects every malformed fixture |
| Determinism | Golden JSON/CSV output diffed in CI; identical across runs and `GOMAXPROCS` |
| Portability | `CGO_ENABLED=0`, standard library only in the library |
| Usability | README example compiles in CI; a tree of 10k nodes builds in < 50 ms |
| Adoption | First consumer live on v0.1; a second internal consumer within six months |

## 9. Open questions

1. **Explicit order vs. sorted-by-name default.** Proposal: nodes with `Order` set sort first by `Order`, the rest by default-language name; confirm.
2. **Slug uniqueness scope.** Proposal: unique per language across the whole tree (needed for flat `/<slug>/` URLs); a `WithSlugScope(Siblings)` option can relax it for path-style URLs.
3. **SKOS mapping of `Parents[0]`.** SKOS has no "primary broader"; proposal: export all parents as `skos:broader` and record the primary in a kit-namespaced property.

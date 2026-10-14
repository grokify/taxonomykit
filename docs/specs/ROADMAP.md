# TaxonomyKit — Roadmap (ROADMAP)

**Initiative:** `INIT-TAXONOMYKIT-001`
**Repository:** `github.com/grokify/taxonomykit`

**Status:** Phases 1–2 planned (Phase 1 = v0.1.0, Phase 2 = v0.2.0)
**RMI slug:** `TAXONOMYKIT`

Phase status is always derived from member RMI statuses, never set directly. Review and execution happen by phase. Commits implementing an RMI carry the git trailer `Refs: RMI-TAXONOMYKIT-NNN`.

## Phase 1 — Core Tree, Codecs & CLI
**Theme:** A validated, deterministic, localized taxonomy tree with polyhierarchy, loadable from tabular or JSON files, usable from Go and the command line.

- [x] `RMI-TAXONOMYKIT-001` Scaffold: go.mod, MIT LICENSE, README, CI (shared Go CI + Lint, CGO_ENABLED=0), CHANGELOG.json
- [x] `RMI-TAXONOMYKIT-002` Types and validation: Node/Label/Labels/Document, every rule with a problem code and an invalid fixture; `Validate` (all problems) and `Build` (first error)
- [x] `RMI-TAXONOMYKIT-003` Tree navigation: indexing, ordering rule with `Order`/name/id and `WithOrderBy`; Roots, Parent(s), Children, Siblings, Lineage(s), Descendants, Leaves, IsLeaf, Depth, LCA, iterators; property tests
- [x] `RMI-TAXONOMYKIT-004` Localization: BCP-47 tag normalization, fallback chain, `Label`, `Breadcrumb` with per-crumb language, `BySlug`, slug-scope option
- [x] `RMI-TAXONOMYKIT-005` Codecs: flat CSV/TSV/PSV reader and writer (header or column mapping, comments, polyhierarchy by repetition), JSON Document with generated and embedded JSON Schema linted in CI
- [ ] `RMI-TAXONOMYKIT-006` CLI `taxonomykit validate|convert|tree|lineage`, CLI tests, CHANGELOG, coverage badge, v0.1.0 release

## Phase 2 — Interchange & Tree Operations
**Theme:** Exchange taxonomies with the world via SKOS and maintain them with library-level operations.

- [ ] `RMI-TAXONOMYKIT-007` skos/: JSON-LD export (ConceptScheme, Concept, prefLabel/altLabel, broader/narrower, topConceptOf, primary-parent property) and import of the same subset; `convert --to skos`
- [ ] `RMI-TAXONOMYKIT-008` Document operations: Subtree, Prune, Merge with conflict report, Diff (added/removed/moved/relabeled); CLI `diff` and `subtree`
- [ ] `RMI-TAXONOMYKIT-009` `validate --strict-lang` translation-gap report and multi-parent report
- [ ] `RMI-TAXONOMYKIT-010` Docs per package and v0.2.0 release

## Post-v0.2 backlog (unphased)

- Multiple schemes (facets) per document
- `skos:related` and mapping relations (`exactMatch`, `closeMatch`) between schemes
- Turtle serialization if a consumer needs it (requires an RDF parser dependency — likely a separate module)
- Materialized-path export helpers for SQL consumers

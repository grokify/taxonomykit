# TaxonomyKit — Implementation Plan (PLAN)

**Status:** Draft
**Date:** 2026-10-14
**Repo:** github.com/grokify/taxonomykit

This plan sequences the work defined in the [PRD](PRD.md) and [TRD](TRD.md). Roadmap items and phase membership are tracked in [ROADMAP.md](ROADMAP.md).

## Guiding constraints

- **Validation first.** The rule table (TRD §4) and its fixtures land before navigation, so every later operation is defined only on valid trees.
- **Determinism from the first commit.** The ordering rule is implemented in `Build` and tested with shuffled input before any accessor exists.
- **Core stays standard-library only.** Codecs and the CLI are separate packages; the schema generator is build-tagged out.
- **Each phase leaves `main` building, tested, lint-clean, with the README example compiling.**
- **Verify dependency versions** (`cobra`, `invopop/jsonschema`) at adoption time.

## Phase 1 — Core tree, codecs, CLI (v0.1.0)

**Goal:** A consumer can load a legacy PSV topic file, get a validated tree with polyhierarchy and two languages, and render breadcrumbs, menus, and slug lookups from it; `taxonomykit validate|convert|tree` work from the command line.

Order of work:

1. Scaffold: `go.mod`, MIT LICENSE, README, CI (shared grokify Go CI + Lint workflows, `CGO_ENABLED=0`), `CHANGELOG.json`.
2. `node.go`, `validate.go`: types, problem codes, every rule with a fixture in `testdata/invalid/`; `Validate` returns all problems, `Build` the first.
3. `build.go`, `tree.go`: indexing (by id, parent→children, child→parents), ordering rule with `Order`/name/id and `WithOrderBy`; `Node Roots Parent Parents Children Siblings Lineage Lineages Descendants Leaves IsLeaf Depth LCA`; `walk.go` iterators; property tests.
4. `locale.go`: tag normalization, fallback chain, `Label`, `Breadcrumb` with per-crumb `Lang`, `BySlug` with slug-scope option.
5. `flat/`: CSV/TSV/PSV reader with header or `Columns`, comment/blank handling, polyhierarchy-by-repetition, writer; golden tests. `schema/`: `Document` JSON, generated schema, embed, `tools.go`, `schemakit lint` in CI.
6. `cmd/taxonomykit`: `validate`, `convert`, `tree`, `lineage`; CLI tests.
7. Release: CHANGELOG, coverage badge, `v0.1.0` after CI; first consumer switches from a local `replace` to the tag.

**Verify:** PRD §8 rows; a two-language, multi-parent fixture mirroring the first consumer's shape round-trips PSV → Document → PSV byte-identically and yields the expected lineages, including both paths for the multi-parent node.

## Phase 2 — Interchange and tree operations (v0.2.0)

**Goal:** Taxonomies move in and out of SKOS, and common maintenance operations are library calls.

Order of work:

1. `skos/`: JSON-LD export with fixed context; import of the same subset; round-trip tests; `convert --to skos`.
2. Tree operations returning new documents: `Subtree(id)`, `Prune(ids)`, `Merge(a, b)` with conflict report, `Diff(a, b)` listing added/removed/moved/relabeled nodes; CLI `diff` and `subtree`.
3. Multiple schemes per document (facets) if a consumer needs it — decision point, not a commitment.
4. Docs page per package; `v0.2.0`.

**Depends on:** Phase 1.

## Risks and mitigations

| Risk | Mitigation |
|------|------------|
| Polyhierarchy makes "the" lineage ambiguous for consumers | Primary parent is explicit and first; `Lineage` is documented as primary-only, `Lineages` as exhaustive |
| Locale fallback hides missing translations | `Crumb.Lang` and `Label` report the language actually used; `validate --strict-lang xx` lists gaps |
| Legacy files with last-wins duplicates change meaning when read as polyhierarchy | `validate` reports every multi-parent node; the consumer reviews once at migration |
| Scope creep toward a CMS or ontology engine | PRD non-goals; new relation types require a consumer and a phase |

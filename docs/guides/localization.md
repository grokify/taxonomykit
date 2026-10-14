# Localization

Every node carries a `Labels` map keyed by language tag, and every lookup
that returns text takes a language.

## Language tags

Tags are BCP-47, normalized to lowercase with hyphens: `en`, `es`, `pt-br`.
`NormalizeLang` applies the same normalization to input (`"EN_us"` →
`"en-us"`), so callers can pass whatever their framework gives them.
`Validate` rejects tags that do not look like BCP-47 (`lang.invalid`).

## The fallback chain

A lookup in a language the node does not have does not fail; it falls back:

```text
pt-br-x  →  pt-br  →  pt  →  <default language>
```

`FallbackChain(lang, defaultLang)` exposes the chain. The default language
is the one every node is guaranteed to have (that is what
`label.default.missing` enforces).

## Lookups

| Call | Returns |
|------|---------|
| `Label(id, lang)` | the `Label`, **the language it was taken from**, and whether the node has any usable label |
| `Name(id, lang)` | the name after fallback, or `""` |
| `Slug(id, lang)` | the slug after fallback, or `""` if no language on the chain has one |
| `Breadcrumb(id, lang)` | the primary lineage as `[]Crumb{ID, Name, Slug, Lang}` |
| `Langs()` | every language that appears in any label |
| `DefaultLang()` | the default |

`Name` and `Slug` fall back independently: a node with a Spanish name but no
Spanish slug gets the Spanish name and the English slug. That is usually what
a site wants — show the translation, keep the URL that exists — and `Label`
tells you when it happened.

## Breadcrumbs that admit their fallbacks

```go
for _, c := range tree.Breadcrumb("patent", "es") {
    if c.Lang != "es" {
        // render with a "not yet translated" hint, or just render
    }
}
```

## Slug lookup

`BySlug(lang, slug)` resolves a URL segment to a node using the **effective**
slug of every node in that language — the slug after fallback. Explicit slugs
in the requested language win over fallback slugs, so a Spanish `derecho`
resolves to `law` while the English `law` does not resolve in Spanish once a
Spanish slug exists. Requests in a language the tree has never seen resolve
through the chain to one it has.

## Slug uniqueness

By default slugs must be unique per language across the whole tree
(`slug.duplicate` otherwise), which is what flat `/<slug>/` URL schemes
need. For path-style URLs (`/<parent>/<child>/`) use
`WithSlugScope(SlugScopeSiblings)` and slugs only need to be unique among
siblings.

# Concepts

## Node

A node is one entry in the taxonomy:

```go
type Node struct {
    ID      string            // canonical, locale-independent identity
    Parents []string          // empty for a root; Parents[0] is the primary parent
    Labels  Labels            // per-language name, slug and alternative names
    Order   *int              // explicit sibling order (optional)
    Attrs   map[string]string // anything else; never interpreted
}

type Label struct {
    Name string   // display name
    Slug string   // URL segment; empty means "no URL in this language"
    Alt  []string // alternative names
}
```

**One id.** Legacy taxonomies tend to accrete several identifiers per node — a
short code, a slug, a sanitized id for a search index. TaxonomyKit has one
canonical `ID` (lowercase, `^[a-z0-9][a-z0-9._-]*$`) and treats slugs as
*labels*: they can differ per language and are never the identity.

**Slugs are never derived.** An empty slug means the node has no URL in that
language. If you want slugs generated from names, do it before `Build`; the
kit will not guess.

## Primary parent and polyhierarchy

A node may have several parents. `Parents[0]` is the **primary** parent, and
it is what single-path operations follow: `Parent`, `Lineage`, `Depth`,
`Breadcrumb`, `LCA`. `Parents` and `Lineages` expose every parent and every
root path. This keeps the common case (one breadcrumb) simple while
preserving the links a single-parent model would have to throw away.

## Ordering

Everything the tree returns is ordered, and the order is defined:

1. nodes with `Order` set, ascending by `Order`;
2. then the rest by the default-language `Name` (case-insensitive, then
   case-sensitive);
3. then by `ID`.

`WithOrderBy(OrderByID)` skips step 2. Because ordering is settled once at
`Build`, the same input gives the same menus, sitemaps and walk order on
every run and on every machine.

## Tree

`Build` validates the nodes and returns a `*Tree`: immutable, indexed, safe
for concurrent reads. Nodes returned by the tree belong to the tree — read
them, don't modify them. Slices returned by the tree are fresh copies.

## Document

```go
type Document struct {
    Version     string // "1"
    DefaultLang string
    Nodes       []Node
}
```

The serializable form: one taxonomy ("scheme" in SKOS terms). `flat.Read`
produces one; `ParseDocument` reads the JSON form strictly; `doc.JSON()`
writes it deterministically; `tree.Document()` exports a tree back. The JSON
Schema for it is generated from these Go types and embedded in the `schema`
package.

## Default language

Every node must have a `Name` in the default language — it anchors ordering
and is the last step of every label lookup. Set it with
`WithDefaultLang("en")` on `Build`/`Validate`, or let a `Document` carry it.

# Navigating a Tree

All navigation is on a `*Tree` from `Build`. Methods take node ids and
return `*Node` values or slices of them; unknown ids return `nil`, `false`
or `-1` rather than panicking. Returned slices are copies; the nodes they
point to belong to the tree.

## Lookup

| Call | Returns |
|------|---------|
| `Node(id)` | the node and whether it exists |
| `Has(id)` | existence |
| `Len()` | number of nodes |
| `Roots()` | nodes without parents, ordered |
| `BySlug(lang, slug)` | the node whose effective slug in `lang` is `slug` (see [Localization](localization.md)) |

## Up

| Call | Returns |
|------|---------|
| `Parent(id)` | the primary parent, or `nil` for a root |
| `Parents(id)` | every parent, primary first |
| `Lineage(id)` | root → … → node through primary parents |
| `Lineages(id)` | every root path; the primary lineage first |
| `Depth(id)` | number of primary-parent edges above the node; roots are 0 |
| `LCA(a, b)` | the deepest node on both primary lineages |

## Across

| Call | Returns |
|------|---------|
| `Children(id)` | direct children, ordered |
| `Siblings(id)` | the other children of the primary parent (other roots for a root), ordered |

## Down

| Call | Returns |
|------|---------|
| `Descendants(id)` | every node below, depth-first pre-order; a node reachable by two paths appears once |
| `Leaves(id)` | childless nodes below |
| `IsLeaf(id)` | the node exists and has no children |

## Walking

```go
tree.Walk(func(n *taxonomykit.Node, depth int) bool {
    fmt.Printf("%*s%s\n", depth*2, "", n.ID)
    return n.ID != "archive" // false skips this node's subtree
})

tree.WalkFrom("science", fn) // the subtree below a node; its children are depth 0

for n, depth := range tree.All() { // Go 1.23 range-over-func
    ...
}
```

Walks are depth-first pre-order from the roots in tree order; a node
reachable by several paths is visited once, on first encounter.

## A rendering example

Breadcrumbs, a sidebar menu and an "also listed in" line for a page about
`mortgage`:

```go
crumbs := tree.Breadcrumb("mortgage", lang)        // primary path
menu   := tree.Children(tree.Parent("mortgage").ID) // the section's items
var also []string
for _, path := range tree.Lineages("mortgage")[1:] { // secondary paths
    also = append(also, tree.Name(path[len(path)-2].ID, lang))
}
```

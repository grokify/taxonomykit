# Validation

`Build` refuses invalid input and returns the first problem; `Validate`
returns all of them so a tool can show the whole list at once. Both take the
same options. Problems are sorted by node id, then code, then language.

```go
if ps := taxonomykit.Validate(nodes, taxonomykit.WithDefaultLang("en")); ps != nil {
    for _, p := range ps {
        fmt.Println(p.Code, p.NodeID, p.Lang, p.Message)
    }
}
```

`Problems` and `Problem` both implement `error`, and `Build`'s error is a
`Problem`, so `errors.As` works:

```go
_, err := taxonomykit.Build(nodes)
var p taxonomykit.Problem
if errors.As(err, &p) && p.Code == taxonomykit.ProblemCycle { ... }
```

## Problem codes

| Code | Meaning | Fix |
|------|---------|-----|
| `id.invalid` | id is empty or not `^[a-z0-9][a-z0-9._-]*$` | lowercase the id; move the display form to a label |
| `id.duplicate` | the same id appears twice | merge the rows (in a flat file, repeated ids are *meant* to merge — this only fires for `[]Node` input) |
| `parent.missing` | a parent id does not exist | add the parent node or fix the id |
| `parent.self` | a node lists itself as a parent | remove it; roots have an empty parent |
| `parent.duplicate` | the same parent listed twice on one node | remove the repeat |
| `cycle` | the node is on a parent cycle; reported for every node on the cycle | break the cycle |
| `root.none` | there are nodes but none without parents | clear the parent of the intended root |
| `label.default.missing` | no non-empty `Name` in the default language | add the label, or change `WithDefaultLang` |
| `slug.duplicate` | two nodes share a slug in one language within the scope | change a slug, or use `WithSlugScope(SlugScopeSiblings)` for path-style URLs |
| `lang.invalid` | a label key is not a BCP-47 tag | use tags like `en`, `es`, `pt-br` |

## What is not validated

- **Translations gaps.** A node missing a label in a non-default language is
  fine; lookups fall back. (A `--strict-lang` report is planned for v0.2.)
- **Multiple parents.** Polyhierarchy is a feature, not a problem. If you
  want to know which nodes have several parents, walk the tree and check
  `len(tree.Parents(id))`.
- **Semantics.** The kit does not know what a category *means*; it checks
  the tree is a tree.

## From the command line

```bash
$ taxonomykit validate topics.csv
parent.missing [b]: parent zzz does not exist
slug.duplicate [a@en]: slug "same" is also used by c
taxonomykit: 2 problem(s)
$ echo $?
1
```

`convert` and `tree` refuse invalid input the same way, so a CI step can be
as simple as `taxonomykit validate topics.psv`.

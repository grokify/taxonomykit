# CLI Reference

```bash
go install github.com/grokify/taxonomykit/cmd/taxonomykit@latest
```

Every command reads one taxonomy file. The input format is inferred from the
file extension (`.csv`, `.tsv`, `.psv`, `.json`); `--from` overrides it, and
`-` reads standard input (with `--from`).

## Common flags

| Flag | Meaning |
|------|---------|
| `--from csv\|tsv\|psv\|json` | input format when it cannot be inferred |
| `--default-lang xx` | default language for flat files (default: the first row's language) |

## `validate <file>`

Prints every structural problem, one per line as `code [id@lang]: message`,
and exits 1 if there are any. On success prints the node count and the
languages found.

## `convert <file> --to csv|tsv|psv|json [-o out]`

Validates, then writes the taxonomy in another format to stdout or `-o`.
Conversion is lossless and deterministic.

## `tree <file> [--lang xx] [--root id] [--depth n]`

Prints an indented listing, one node per line as `Name (id) /slug`, names
and slugs resolved in `--lang` with fallback. `--root` restricts the output
to the subtree below a node (its children at depth 0); `--depth` limits how
deep to print.

## `lineage <file> <id> [--lang xx] [--all]`

Prints the path from the root to the node as `Name (id) > Name (id) > …`.
With `--all`, prints every path of a multi-parent node, primary first.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | success |
| 1 | invalid taxonomy, unreadable file, unknown node, or bad arguments; the reason is on stderr |

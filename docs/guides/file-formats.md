# File Formats

## Flat files: CSV, TSV, PSV

The `flat` package reads and writes delimited text with **one row per node
and language**.

| Format | Delimiter | Notes |
|--------|-----------|-------|
| CSV | `,` | RFC 4180 quoting via `encoding/csv` |
| TSV | tab | no quoting; a cell may not contain a tab |
| PSV | `\|` | cells are trimmed of surrounding whitespace; a cell may not contain a pipe |

Blank lines and lines starting with `#` are skipped in every format.

### Columns

With a header row, column names are matched case-insensitively:

| Column | Role | Required |
|--------|------|----------|
| `lang` (or `language`) | language tag of this row's label; defaults to the document default | no |
| `parent` (or `parentid`, `parent_id`) | parent id; empty for a root | no |
| `id` | node id | **yes** |
| `slug` | URL segment in this language | no |
| `name` (or `display`, `label`) | display name in this language | **yes** |
| `alt` | alternative names, `;`-separated | no |
| `order` | explicit sibling order (integer) | no |
| anything else | stored in `Node.Attrs` under the header name | no |

Without a header row, pass a `flat.Columns` with zero-based positions (`-1`
for absent roles).

### How rows merge

Rows are grouped by `id`; nodes keep first-appearance order.

- The first row for an id sets `Parents[0]` from its parent cell.
- A later row whose parent cell names a **different** non-empty id **adds**
  that parent. This is how a legacy single-parent file that listed a node
  twice becomes a multi-parent node instead of silently keeping the last
  parent.
- A later row with an empty parent cell, or repeating an existing parent,
  changes nothing about parents — so language rows can leave the parent
  blank.
- Each row sets the `Name`, `Slug` and `Alt` for its `lang`.
- `order` is taken from the first row that has one.

```text
lang | parent     | id       | slug     | name
en   | finance    | mortgage | mortgage | Mortgage     ← primary parent
en   | realestate | mortgage | mortgage | Mortgage     ← adds a second parent
es   |            | mortgage | hipoteca | Hipoteca     ← adds a label only
```

### Document default language

`flat.Options.DefaultLang` sets it; otherwise the first data row's language
is used (or `en` if the file has no `lang` column).

### Writing

`flat.Write` emits a header and one row per node and language, nodes in
document order, languages with the default first then sorted, and the first
language row repeated once per parent. Output is deterministic, so a read →
write → read round trip reproduces the same document and the same bytes.

## JSON document

```json
{
  "version": "1",
  "defaultLang": "en",
  "nodes": [
    { "id": "all", "labels": { "en": { "name": "All", "slug": "all" } } },
    { "id": "law", "parents": ["all"],
      "labels": { "en": { "name": "Law", "slug": "law" },
                  "es": { "name": "Derecho", "slug": "derecho" } } },
    { "id": "mortgage", "parents": ["finance", "realestate"], "order": 1,
      "labels": { "en": { "name": "Mortgage", "slug": "mortgage", "alt": ["home loan"] } },
      "attrs": { "legacyCode": "mort" } }
  ]
}
```

- `taxonomykit.ParseDocument` is strict: unknown fields are rejected and the
  version must be `"1"`.
- `doc.JSON()` writes two-space-indented, key-ordered output with a trailing
  newline, so the file diffs cleanly in version control.
- The JSON Schema is in the `schema` package (`schema.Document()` returns the
  bytes; `schema.ID` is its `$id`). It is generated from the Go types with
  `go generate ./schema` and committed, so the types stay the source of
  truth.

## Choosing a format

Keep the taxonomy in whichever form your team edits most comfortably — PSV
is pleasant to read and diff by hand; CSV opens in a spreadsheet — and use
JSON as the interchange form between tools. `taxonomykit convert` moves
between them losslessly.

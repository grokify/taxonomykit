# Quick Start

This walks through the whole kit on a small topic tree: a file, a validated
tree, breadcrumbs in two languages, and the CLI.

## 1. Write the taxonomy as a file

One row per node and language. The parent column names the parent's id;
roots have an empty parent.

```text title="topics.psv"
lang | parent     | id         | slug       | name
en   |            | all        | all        | All
en   | all        | business   | business   | Business
en   | all        | law        | law        | Law
es   |            | law        | derecho    | Derecho
en   | law        | patent     | patent     | Patent
en   | business   | finance    | finance    | Finance
en   | business   | realestate | realestate | Real Estate
en   | finance    | mortgage   | mortgage   | Mortgage
en   | realestate | mortgage   | mortgage   | Mortgage
```

Two things to notice:

- `law` has a second row in Spanish. It adds a label; it does not change
  the parent (the parent cell is empty, and would be ignored even if it
  repeated `all`).
- `mortgage` appears under two parents. It becomes a node with two parents;
  the first row seen (`finance`) is the **primary** parent.

## 2. Validate it

```bash
$ taxonomykit validate topics.psv
ok: 7 nodes, languages en,es
```

A broken file lists every problem and exits 1:

```bash
$ taxonomykit validate broken.csv
parent.missing [b]: parent zzz does not exist
slug.duplicate [a@en]: slug "same" is also used by c
taxonomykit: 2 problem(s)
```

## 3. Load it in Go

```go
f, err := os.Open("topics.psv")
if err != nil {
    return err
}
defer f.Close()

doc, err := flat.Read(f, flat.PSV, flat.Options{})
if err != nil {
    return err
}
tree, err := doc.Build() // uses doc.DefaultLang, "en" here (the first row's language)
if err != nil {
    return err // the first taxonomykit.Problem
}
```

## 4. Ask it questions

```go
tree.Lineage("mortgage")   // all › business › finance › mortgage      (primary parents)
tree.Lineages("mortgage")  // the path above, plus all › business › realestate › mortgage
tree.Children("all")       // business, law   (ordered by name)
tree.Siblings("patent")    // none — patent is law's only child
tree.IsLeaf("patent")      // true
tree.Leaves("business")    // mortgage

for _, c := range tree.Breadcrumb("patent", "es") {
    fmt.Printf("%s (/%s) [%s]  ", c.Name, c.Slug, c.Lang)
}
// All (/all) [en]  Derecho (/derecho) [es]  Patent (/patent) [en]

n, _ := tree.BySlug("es", "derecho") // n.ID == "law"
```

Every crumb says which language it was resolved in, so `[en]` on `Patent`
tells you there is no Spanish label yet.

## 5. Convert and print

```bash
taxonomykit convert topics.psv --to json -o topics.json
taxonomykit tree topics.json --lang es
taxonomykit lineage topics.json mortgage --all
```

```text
All (all) /all
  Business (business) /business
    Finance (finance) /finance
      Mortgage (mortgage) /mortgage
    Real Estate (realestate) /realestate
  Derecho (law) /derecho
    Patent (patent) /patent
```

The JSON file is the canonical interchange form; see
[File Formats](file-formats.md).

# Installation

## Library

```bash
go get github.com/grokify/taxonomykit
```

```go
import (
    "github.com/grokify/taxonomykit"      // tree, validation, localization
    "github.com/grokify/taxonomykit/flat" // CSV / TSV / PSV codec
    "github.com/grokify/taxonomykit/schema" // embedded JSON Schema for the document format
)
```

The library packages depend only on the Go standard library and build with
`CGO_ENABLED=0`. Go 1.26 or later is required.

## CLI

```bash
go install github.com/grokify/taxonomykit/cmd/taxonomykit@latest
taxonomykit --help
```

The CLI is a thin layer over the library; everything it does is available
as a function call.

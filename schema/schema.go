// Package schema embeds the JSON Schema for the taxonomykit Document format.
// The schema is generated from the Go types by schema/gen (go generate) and
// committed alongside them.
package schema

import _ "embed"

//go:generate go run gen/main.go

//go:embed document.schema.json
var document []byte

// Document returns a copy of the Document JSON Schema.
func Document() []byte { return append([]byte(nil), document...) }

// ID is the schema's "$id".
const ID = "https://github.com/grokify/taxonomykit/schema/document.schema.json"

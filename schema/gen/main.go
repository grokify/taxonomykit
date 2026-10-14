//go:build ignore

// Command gen writes document.schema.json from the taxonomykit Go types.
// Run from the schema directory via go generate.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/grokify/taxonomykit"
	"github.com/grokify/taxonomykit/schema"
	"github.com/invopop/jsonschema"
)

func main() {
	r := &jsonschema.Reflector{}
	if err := r.AddGoComments("github.com/grokify/taxonomykit", "../"); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
	s := r.Reflect(&taxonomykit.Document{})
	s.ID = jsonschema.ID(schema.ID)
	s.Title = "TaxonomyKit Document"
	out, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
	if err := os.WriteFile("document.schema.json", append(out, '\n'), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

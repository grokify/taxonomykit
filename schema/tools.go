//go:build tools

package schema

// Keeps the schema generator's dependency in go.mod; gen/main.go is build-tagged
// out, so go mod tidy would otherwise drop it.
import _ "github.com/invopop/jsonschema"

package taxonomykit

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// NewDocument wraps nodes in a Document with the current version.
func NewDocument(defaultLang string, nodes []Node) *Document {
	return &Document{Version: DocumentVersion, DefaultLang: NormalizeLang(defaultLang), Nodes: nodes}
}

// ParseDocument decodes the JSON document format strictly: unknown fields are
// rejected and the version must be supported.
func ParseDocument(data []byte) (*Document, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var d Document
	if err := dec.Decode(&d); err != nil {
		return nil, fmt.Errorf("taxonomykit: parse document: %w", err)
	}
	if d.Version != DocumentVersion {
		return nil, fmt.Errorf("taxonomykit: unsupported document version %q (want %q)", d.Version, DocumentVersion)
	}
	d.DefaultLang = NormalizeLang(d.DefaultLang)
	return &d, nil
}

// JSON encodes the document deterministically with two-space indentation
// and a trailing newline.
func (d *Document) JSON() ([]byte, error) {
	out, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("taxonomykit: encode document: %w", err)
	}
	return append(out, '\n'), nil
}

// Validate runs Validate on the document's nodes with its default language.
func (d *Document) Validate(opts ...Option) Problems {
	return Validate(d.Nodes, append([]Option{WithDefaultLang(d.DefaultLang)}, opts...)...)
}

// Build runs Build on the document's nodes with its default language.
func (d *Document) Build(opts ...Option) (*Tree, error) {
	return Build(d.Nodes, append([]Option{WithDefaultLang(d.DefaultLang)}, opts...)...)
}

// Document returns the tree's nodes as a Document, in the tree's input order
// with normalized language tags.
func (t *Tree) Document() *Document {
	nodes := make([]Node, len(t.nodes))
	for i := range t.nodes {
		nodes[i] = t.nodes[i].clone()
	}
	return NewDocument(t.opts.defaultLang, nodes)
}

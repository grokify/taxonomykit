// Package taxonomykit builds validated, deterministic, localized taxonomy
// trees from a flat list of nodes and answers the structural questions
// applications ask of them: lineage, children, siblings, leaves, breadcrumbs
// and slug lookup. It owns structure only; the content of a taxonomy stays
// with the caller.
package taxonomykit

// DocumentVersion is the current Document format version.
const DocumentVersion = "1"

// Node is one entry in a taxonomy. ID is the canonical, locale-independent
// identity; everything else is a label or a relation.
type Node struct {
	// ID is unique within a document and matches ^[a-z0-9][a-z0-9._-]*$.
	ID string `json:"id"`
	// Parents lists parent IDs; empty for a root. Parents[0] is the primary
	// parent, used for Lineage and Breadcrumb.
	Parents []string `json:"parents,omitempty"`
	// Labels holds the per-language name, slug and alternative names, keyed by
	// a lowercase BCP-47 tag such as "en" or "pt-br".
	Labels Labels `json:"labels"`
	// Order is an explicit sibling order. Nodes with Order set sort before
	// nodes without, ascending; nil nodes sort by default-language name.
	Order *int `json:"order,omitempty"`
	// Attrs carries caller-defined data the kit never interprets.
	Attrs map[string]string `json:"attrs,omitempty"`
}

// Label is a node's presentation in one language.
type Label struct {
	Name string `json:"name"`
	// Slug is the URL segment for this language. Empty means the node has no
	// URL in this language; a slug is never derived from Name.
	Slug string   `json:"slug,omitempty"`
	Alt  []string `json:"alt,omitempty"`
}

// Labels maps a lowercase BCP-47 language tag to a Label.
type Labels map[string]Label

// Document is the serializable form of a taxonomy: one scheme.
type Document struct {
	Version     string `json:"version"`
	DefaultLang string `json:"defaultLang"`
	Nodes       []Node `json:"nodes"`
}

// Crumb is one element of a localized breadcrumb.
type Crumb struct {
	ID   string
	Name string
	Slug string
	// Lang is the language the Name and Slug were taken from after fallback.
	Lang string
}

// clone returns a deep copy so a Tree never aliases caller memory.
func (n Node) clone() Node {
	c := Node{ID: n.ID}
	if n.Parents != nil {
		c.Parents = append([]string(nil), n.Parents...)
	}
	if n.Labels != nil {
		c.Labels = make(Labels, len(n.Labels))
		for lang, l := range n.Labels {
			cl := Label{Name: l.Name, Slug: l.Slug}
			if l.Alt != nil {
				cl.Alt = append([]string(nil), l.Alt...)
			}
			c.Labels[lang] = cl
		}
	}
	if n.Order != nil {
		o := *n.Order
		c.Order = &o
	}
	if n.Attrs != nil {
		c.Attrs = make(map[string]string, len(n.Attrs))
		for k, v := range n.Attrs {
			c.Attrs[k] = v
		}
	}
	return c
}

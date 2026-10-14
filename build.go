package taxonomykit

import (
	"sort"
	"strings"
)

// OrderBy selects the tie-break used for nodes without an explicit Order.
type OrderBy int

const (
	// OrderByName sorts by default-language name (case-insensitive, then
	// case-sensitive), then by ID.
	OrderByName OrderBy = iota
	// OrderByID sorts by ID only.
	OrderByID
)

type options struct {
	defaultLang string
	slugScope   SlugScope
	orderBy     OrderBy
}

// Option configures Build and Validate.
type Option func(*options)

// WithDefaultLang sets the language every node must have a name in and the
// last step of the label fallback chain. The default is "en".
func WithDefaultLang(lang string) Option {
	return func(o *options) { o.defaultLang = NormalizeLang(lang) }
}

// WithSlugScope sets where slugs must be unique. The default is SlugScopeTree.
func WithSlugScope(s SlugScope) Option { return func(o *options) { o.slugScope = s } }

// WithOrderBy sets the tie-break for nodes without an explicit Order.
func WithOrderBy(b OrderBy) Option { return func(o *options) { o.orderBy = b } }

func applyOptions(opts []Option) options {
	o := options{defaultLang: "en"}
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// Tree is a validated, immutable, indexed taxonomy. It is safe for
// concurrent use. Nodes returned by its methods belong to the tree and must
// not be modified.
type Tree struct {
	opts     options
	nodes    []Node // tree-owned copies, in input order
	byID     map[string]*Node
	roots    []*Node
	children map[string][]*Node
	parents  map[string][]*Node
	depth    map[string]int
	langs    []string
	slugs    map[string]map[string]*Node // lang → effective slug → node
}

// Build validates nodes and returns the indexed tree, or the first Problem
// found. Use Validate to see every problem at once.
func Build(nodes []Node, opts ...Option) (*Tree, error) {
	if ps := Validate(nodes, opts...); len(ps) > 0 {
		return nil, ps[0]
	}
	t := &Tree{
		opts:     applyOptions(opts),
		nodes:    make([]Node, len(nodes)),
		byID:     make(map[string]*Node, len(nodes)),
		children: map[string][]*Node{},
		parents:  map[string][]*Node{},
		depth:    make(map[string]int, len(nodes)),
	}
	langSet := map[string]bool{}
	for i, n := range nodes {
		c := n.clone()
		if c.Labels != nil {
			norm := make(Labels, len(c.Labels))
			for lang, l := range c.Labels {
				norm[NormalizeLang(lang)] = l
			}
			c.Labels = norm
			for lang := range norm {
				langSet[lang] = true
			}
		}
		t.nodes[i] = c
		t.byID[c.ID] = &t.nodes[i]
	}
	for i := range t.nodes {
		n := &t.nodes[i]
		if len(n.Parents) == 0 {
			t.roots = append(t.roots, n)
			continue
		}
		for _, pid := range n.Parents {
			p := t.byID[pid]
			t.children[pid] = append(t.children[pid], n)
			t.parents[n.ID] = append(t.parents[n.ID], p)
		}
	}
	t.sortNodes(t.roots)
	for id := range t.children {
		t.sortNodes(t.children[id])
	}
	for i := range t.nodes {
		t.depth[t.nodes[i].ID] = len(t.Lineage(t.nodes[i].ID)) - 1
	}
	for lang := range langSet {
		t.langs = append(t.langs, lang)
	}
	sort.Strings(t.langs)
	t.slugs = map[string]map[string]*Node{}
	for _, lang := range t.langs {
		t.slugs[lang] = t.effectiveSlugIndex(lang)
	}
	return t, nil
}

// MustBuild is Build for static, known-good input; it panics on error.
func MustBuild(nodes []Node, opts ...Option) *Tree {
	t, err := Build(nodes, opts...)
	if err != nil {
		panic("taxonomykit: " + err.Error())
	}
	return t
}

// DefaultLang is the language used as the last fallback and for ordering.
func (t *Tree) DefaultLang() string { return t.opts.defaultLang }

// Langs lists every language tag that appears in any label, sorted.
func (t *Tree) Langs() []string { return append([]string(nil), t.langs...) }

// Len is the number of nodes.
func (t *Tree) Len() int { return len(t.nodes) }

// sortNodes orders siblings in place: explicit Order first ascending, then
// by name (or ID) per the OrderBy option, then by ID.
func (t *Tree) sortNodes(ns []*Node) {
	sort.SliceStable(ns, func(i, j int) bool {
		a, b := ns[i], ns[j]
		switch {
		case a.Order != nil && b.Order != nil:
			if *a.Order != *b.Order {
				return *a.Order < *b.Order
			}
		case a.Order != nil:
			return true
		case b.Order != nil:
			return false
		}
		if t.opts.orderBy == OrderByName {
			an, bn := t.defaultName(a), t.defaultName(b)
			if la, lb := strings.ToLower(an), strings.ToLower(bn); la != lb {
				return la < lb
			}
			if an != bn {
				return an < bn
			}
		}
		return a.ID < b.ID
	})
}

func (t *Tree) defaultName(n *Node) string {
	return n.Labels[t.opts.defaultLang].Name
}

// effectiveSlugIndex maps each node's resolved slug in lang (after fallback)
// to the node. An explicit slug in lang wins over a fallback slug.
func (t *Tree) effectiveSlugIndex(lang string) map[string]*Node {
	idx := map[string]*Node{}
	explicit := map[string]bool{}
	for i := range t.nodes {
		n := &t.nodes[i]
		l, used, ok := t.Label(n.ID, lang)
		if !ok || l.Slug == "" {
			continue
		}
		if used == lang {
			idx[l.Slug] = n
			explicit[l.Slug] = true
			continue
		}
		if !explicit[l.Slug] {
			if _, taken := idx[l.Slug]; !taken {
				idx[l.Slug] = n
			}
		}
	}
	return idx
}

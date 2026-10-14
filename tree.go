package taxonomykit

// Node returns the node with the given ID.
func (t *Tree) Node(id string) (*Node, bool) {
	n, ok := t.byID[id]
	return n, ok
}

// Has reports whether a node with the given ID exists.
func (t *Tree) Has(id string) bool {
	_, ok := t.byID[id]
	return ok
}

// Roots returns the nodes without parents, ordered.
func (t *Tree) Roots() []*Node { return copyNodes(t.roots) }

// Parent returns the primary parent, or nil for a root or unknown ID.
func (t *Tree) Parent(id string) *Node {
	ps := t.parents[id]
	if len(ps) == 0 {
		return nil
	}
	return ps[0]
}

// Parents returns every parent in declaration order (primary first).
func (t *Tree) Parents(id string) []*Node { return copyNodes(t.parents[id]) }

// Children returns the direct children, ordered.
func (t *Tree) Children(id string) []*Node { return copyNodes(t.children[id]) }

// Siblings returns the other children of the primary parent, ordered. For a
// root it returns the other roots.
func (t *Tree) Siblings(id string) []*Node {
	n, ok := t.byID[id]
	if !ok {
		return nil
	}
	var pool []*Node
	if p := t.Parent(id); p != nil {
		pool = t.children[p.ID]
	} else {
		pool = t.roots
	}
	out := make([]*Node, 0, len(pool))
	for _, s := range pool {
		if s != n {
			out = append(out, s)
		}
	}
	return out
}

// Lineage returns the path from a root to the node through primary parents,
// root first. It is nil for an unknown ID.
func (t *Tree) Lineage(id string) []*Node {
	n, ok := t.byID[id]
	if !ok {
		return nil
	}
	var rev []*Node
	for n != nil {
		rev = append(rev, n)
		n = t.Parent(n.ID)
	}
	out := make([]*Node, len(rev))
	for i, x := range rev {
		out[len(rev)-1-i] = x
	}
	return out
}

// Lineages returns every root-to-node path through any parent, each root
// first, with the primary lineage first and the rest in sibling order of
// their first divergence.
func (t *Tree) Lineages(id string) [][]*Node {
	n, ok := t.byID[id]
	if !ok {
		return nil
	}
	var out [][]*Node
	var walk func(cur *Node, path []*Node)
	walk = func(cur *Node, path []*Node) {
		path = append([]*Node{cur}, path...)
		ps := t.parents[cur.ID]
		if len(ps) == 0 {
			out = append(out, path)
			return
		}
		for _, p := range ps {
			walk(p, path)
		}
	}
	walk(n, nil)
	return out
}

// Descendants returns every node below id in depth-first pre-order; a node
// reachable by several paths appears once.
func (t *Tree) Descendants(id string) []*Node {
	if _, ok := t.byID[id]; !ok {
		return nil
	}
	var out []*Node
	seen := map[*Node]bool{}
	var walk func(cur string)
	walk = func(cur string) {
		for _, c := range t.children[cur] {
			if seen[c] {
				continue
			}
			seen[c] = true
			out = append(out, c)
			walk(c.ID)
		}
	}
	walk(id)
	return out
}

// Leaves returns the childless nodes below id in depth-first order. For a
// leaf it returns nil.
func (t *Tree) Leaves(id string) []*Node {
	var out []*Node
	for _, d := range t.Descendants(id) {
		if len(t.children[d.ID]) == 0 {
			out = append(out, d)
		}
	}
	return out
}

// IsLeaf reports whether the node exists and has no children.
func (t *Tree) IsLeaf(id string) bool {
	_, ok := t.byID[id]
	return ok && len(t.children[id]) == 0
}

// Depth is the number of primary-parent edges above the node; roots are 0.
// It is -1 for an unknown ID.
func (t *Tree) Depth(id string) int {
	d, ok := t.depth[id]
	if !ok {
		return -1
	}
	return d
}

// LCA returns the deepest node on both primary lineages, or nil if the IDs
// are unknown or lie under different roots.
func (t *Tree) LCA(a, b string) *Node {
	la, lb := t.Lineage(a), t.Lineage(b)
	var lca *Node
	for i := 0; i < len(la) && i < len(lb) && la[i] == lb[i]; i++ {
		lca = la[i]
	}
	return lca
}

func copyNodes(ns []*Node) []*Node {
	if ns == nil {
		return nil
	}
	return append([]*Node(nil), ns...)
}

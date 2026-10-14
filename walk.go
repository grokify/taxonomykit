package taxonomykit

import "iter"

// Walk visits every node reachable from the roots in depth-first pre-order
// with its depth. Returning false from fn skips the node's subtree. A node
// reachable by several paths is visited once, on its first encounter.
func (t *Tree) Walk(fn func(n *Node, depth int) bool) {
	t.walkFrom(t.roots, 0, fn)
}

// WalkFrom is Walk starting below the given node, whose children are depth 0.
func (t *Tree) WalkFrom(id string, fn func(n *Node, depth int) bool) {
	t.walkFrom(t.children[id], 0, fn)
}

func (t *Tree) walkFrom(start []*Node, depth int, fn func(n *Node, depth int) bool) {
	seen := map[*Node]bool{}
	var walk func(ns []*Node, d int) bool
	walk = func(ns []*Node, d int) bool {
		for _, n := range ns {
			if seen[n] {
				continue
			}
			seen[n] = true
			if !fn(n, d) {
				continue
			}
			if !walk(t.children[n.ID], d+1) {
				return false
			}
		}
		return true
	}
	walk(start, depth)
}

// All iterates over every node in depth-first pre-order from the roots,
// yielding the node and its depth.
func (t *Tree) All() iter.Seq2[*Node, int] {
	return func(yield func(*Node, int) bool) {
		stop := false
		t.Walk(func(n *Node, d int) bool {
			if stop {
				return false
			}
			if !yield(n, d) {
				stop = true
				return false
			}
			return true
		})
	}
}

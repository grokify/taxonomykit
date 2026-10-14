package taxonomykit

import (
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestNavigation(t *testing.T) {
	tr := MustBuild(fixture())

	if got := ids(tr.Roots()); !reflect.DeepEqual(got, []string{"all"}) {
		t.Fatalf("roots: %v", got)
	}
	// technology (Order 0) and science (Order 1) first, then by name.
	if got := ids(tr.Children("all")); !reflect.DeepEqual(got, []string{"technology", "science", "business", "law"}) {
		t.Fatalf("children of all: %v", got)
	}
	if got := ids(tr.Siblings("patent")); !reflect.DeepEqual(got, []string{"trademark"}) {
		t.Fatalf("siblings of patent: %v", got)
	}
	if got := ids(tr.Siblings("all")); len(got) != 0 {
		t.Fatalf("a lone root has no siblings: %v", got)
	}
	if got := ids(tr.Lineage("mortgage")); !reflect.DeepEqual(got, []string{"all", "business", "finance", "personal-finance", "mortgage"}) {
		t.Fatalf("primary lineage: %v", got)
	}
	lins := tr.Lineages("mortgage")
	if len(lins) != 2 {
		t.Fatalf("want 2 lineages, got %d", len(lins))
	}
	if got := ids(lins[1]); !reflect.DeepEqual(got, []string{"all", "business", "realestate", "mortgage"}) {
		t.Fatalf("secondary lineage: %v", got)
	}
	if p := tr.Parent("mortgage"); p == nil || p.ID != "personal-finance" {
		t.Fatalf("primary parent: %v", p)
	}
	if got := ids(tr.Parents("mortgage")); !reflect.DeepEqual(got, []string{"personal-finance", "realestate"}) {
		t.Fatalf("parents: %v", got)
	}
	if got := ids(tr.Descendants("business")); !reflect.DeepEqual(got, []string{"finance", "personal-finance", "mortgage", "realestate"}) {
		t.Fatalf("descendants (mortgage once): %v", got)
	}
	if got := ids(tr.Leaves("business")); !reflect.DeepEqual(got, []string{"mortgage"}) {
		t.Fatalf("leaves: %v", got)
	}
	if !tr.IsLeaf("water") || tr.IsLeaf("science") || tr.IsLeaf("nope") {
		t.Fatal("IsLeaf")
	}
	if tr.Depth("all") != 0 || tr.Depth("mortgage") != 4 || tr.Depth("nope") != -1 {
		t.Fatalf("depth: %d %d %d", tr.Depth("all"), tr.Depth("mortgage"), tr.Depth("nope"))
	}
	if lca := tr.LCA("mortgage", "patent"); lca == nil || lca.ID != "all" {
		t.Fatalf("lca: %v", lca)
	}
	if lca := tr.LCA("mortgage", "finance"); lca == nil || lca.ID != "finance" {
		t.Fatalf("lca ancestor: %v", lca)
	}
	if tr.Lineage("nope") != nil || tr.Children("nope") != nil || tr.Parent("nope") != nil {
		t.Fatal("unknown ids return nil")
	}
	if tr.Len() != 12 || !tr.Has("water") {
		t.Fatal("Len/Has")
	}
}

func TestOrderIsDeterministicUnderShuffle(t *testing.T) {
	want := MustBuild(fixture())
	var wantWalk []string
	want.Walk(func(n *Node, d int) bool { wantWalk = append(wantWalk, n.ID); return true })

	nodes := fixture()
	r := rand.New(rand.NewPCG(7, 11)) //nolint:gosec // seeded shuffle; determinism is the point of the test
	for i := 0; i < 20; i++ {
		r.Shuffle(len(nodes), func(a, b int) { nodes[a], nodes[b] = nodes[b], nodes[a] })
		got := MustBuild(nodes)
		var gotWalk []string
		got.Walk(func(n *Node, d int) bool { gotWalk = append(gotWalk, n.ID); return true })
		if !reflect.DeepEqual(gotWalk, wantWalk) {
			t.Fatalf("walk order changed with input order:\n want %v\n got  %v", wantWalk, gotWalk)
		}
	}
}

func TestOrderByID(t *testing.T) {
	tr := MustBuild(fixture(), WithOrderBy(OrderByID))
	if got := ids(tr.Children("all")); !reflect.DeepEqual(got, []string{"technology", "science", "business", "law"}) {
		t.Fatalf("explicit Order still wins; rest by id: %v", got)
	}
	if got := ids(tr.Children("business")); !reflect.DeepEqual(got, []string{"finance", "realestate"}) {
		t.Fatalf("by id: %v", got)
	}
}

func TestWalkAndAll(t *testing.T) {
	tr := MustBuild(fixture())
	var depths []int
	tr.Walk(func(n *Node, d int) bool {
		depths = append(depths, d)
		return n.ID != "business" // skip the business subtree
	})
	for _, d := range depths {
		if d > 2 {
			t.Fatalf("depth %d reached although business subtree was skipped", d)
		}
	}
	count := 0
	for n, d := range tr.All() {
		if n.ID == "all" && d != 0 {
			t.Fatal("root depth")
		}
		count++
		if count == 3 {
			break
		}
	}
	if count != 3 {
		t.Fatalf("iterator break: %d", count)
	}
	var below []string
	tr.WalkFrom("law", func(n *Node, d int) bool { below = append(below, n.ID); return true })
	if !reflect.DeepEqual(below, []string{"patent", "trademark"}) {
		t.Fatalf("WalkFrom: %v", below)
	}
}

func TestTreeDoesNotAliasInput(t *testing.T) {
	nodes := fixture()
	tr := MustBuild(nodes)
	nodes[0].Labels["en"] = Label{Name: "Changed"}
	nodes[1].Parents[0] = "changed"
	if tr.Name("all", "en") != "All" || tr.Parent("business").ID != "all" {
		t.Fatal("tree must copy its input")
	}
	cs := tr.Children("all")
	cs[0] = nil
	if tr.Children("all")[0] == nil {
		t.Fatal("returned slices must be copies")
	}
}

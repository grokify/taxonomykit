package taxonomykit_test

import (
	"fmt"
	"strings"

	"github.com/grokify/taxonomykit"
)

// This example is the README example; it compiles and runs in CI.
func ExampleBuild() {
	tree, err := taxonomykit.Build([]taxonomykit.Node{
		{ID: "all", Labels: taxonomykit.Labels{"en": {Name: "All", Slug: "all"}}},
		{ID: "law", Parents: []string{"all"}, Labels: taxonomykit.Labels{
			"en": {Name: "Law", Slug: "law"},
			"es": {Name: "Derecho", Slug: "derecho"},
		}},
		{ID: "patent", Parents: []string{"law"}, Labels: taxonomykit.Labels{"en": {Name: "Patent", Slug: "patent"}}},
	}, taxonomykit.WithDefaultLang("en"))
	if err != nil {
		panic(err) // cycles, orphans, duplicate ids or slugs, missing default labels
	}

	var lineage []string
	for _, n := range tree.Lineage("patent") {
		lineage = append(lineage, n.ID)
	}
	fmt.Println(strings.Join(lineage, " > "))

	var crumbs []string
	for _, c := range tree.Breadcrumb("patent", "es") {
		crumbs = append(crumbs, fmt.Sprintf("%s(/%s,%s)", c.Name, c.Slug, c.Lang))
	}
	fmt.Println(strings.Join(crumbs, " > "))

	fmt.Println(tree.IsLeaf("patent"), len(tree.Children("all")), len(tree.Siblings("patent")))
	n, _ := tree.BySlug("es", "derecho")
	fmt.Println(n.ID)
	// Output:
	// all > law > patent
	// All(/all,en) > Derecho(/derecho,es) > Patent(/patent,en)
	// true 1 0
	// law
}

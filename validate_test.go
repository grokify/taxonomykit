package taxonomykit

import (
	"errors"
	"testing"
)

func TestValidateFixtureIsClean(t *testing.T) {
	if ps := Validate(fixture()); ps != nil {
		t.Fatalf("expected no problems, got %v", ps)
	}
}

func TestValidateProblems(t *testing.T) {
	en := func(name, slug string) Labels { return Labels{"en": {Name: name, Slug: slug}} }
	cases := []struct {
		name  string
		nodes []Node
		code  string
		node  string
	}{
		{"id invalid", []Node{{ID: "Bad ID", Labels: en("x", "x")}}, ProblemIDInvalid, "Bad ID"},
		{"id duplicate", []Node{{ID: "a", Labels: en("a", "a")}, {ID: "a", Labels: en("a", "a2")}}, ProblemIDDuplicate, "a"},
		{"parent missing", []Node{{ID: "a", Labels: en("a", "a")}, {ID: "b", Parents: []string{"zz"}, Labels: en("b", "b")}}, ProblemParentMissing, "b"},
		{"parent self", []Node{{ID: "a", Labels: en("a", "a")}, {ID: "b", Parents: []string{"b"}, Labels: en("b", "b")}}, ProblemParentSelf, "b"},
		{"parent duplicate", []Node{{ID: "a", Labels: en("a", "a")}, {ID: "b", Parents: []string{"a", "a"}, Labels: en("b", "b")}}, ProblemParentDuplicate, "b"},
		{"cycle", []Node{{ID: "r", Labels: en("r", "r")}, {ID: "a", Parents: []string{"b"}, Labels: en("a", "a")}, {ID: "b", Parents: []string{"a"}, Labels: en("b", "b")}}, ProblemCycle, "a"},
		{"root none", []Node{{ID: "a", Parents: []string{"b"}, Labels: en("a", "a")}, {ID: "b", Parents: []string{"a"}, Labels: en("b", "b")}}, ProblemRootNone, ""},
		{"label default missing", []Node{{ID: "a", Labels: Labels{"es": {Name: "a"}}}}, ProblemLabelDefaultMissing, "a"},
		{"label default empty name", []Node{{ID: "a", Labels: Labels{"en": {Name: ""}}}}, ProblemLabelDefaultMissing, "a"},
		{"slug duplicate", []Node{{ID: "a", Labels: en("a", "same")}, {ID: "b", Labels: en("b", "same")}}, ProblemSlugDuplicate, "a"},
		{"lang invalid", []Node{{ID: "a", Labels: Labels{"en": {Name: "a"}, "english!": {Name: "a"}}}}, ProblemLangInvalid, "a"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ps := Validate(tc.nodes)
			for _, p := range ps {
				if p.Code == tc.code && p.NodeID == tc.node {
					return
				}
			}
			t.Fatalf("want problem %s on %q, got %v", tc.code, tc.node, ps)
		})
	}
}

func TestValidateSlugScopeSiblings(t *testing.T) {
	en := func(name, slug string) Labels { return Labels{"en": {Name: name, Slug: slug}} }
	nodes := []Node{
		{ID: "r", Labels: en("r", "r")},
		{ID: "a", Parents: []string{"r"}, Labels: en("a", "a")},
		{ID: "b", Parents: []string{"r"}, Labels: en("b", "b")},
		{ID: "a-x", Parents: []string{"a"}, Labels: en("x", "x")},
		{ID: "b-x", Parents: []string{"b"}, Labels: en("x", "x")},
	}
	if ps := Validate(nodes); ps == nil {
		t.Fatal("tree scope should reject the repeated slug")
	}
	if ps := Validate(nodes, WithSlugScope(SlugScopeSiblings)); ps != nil {
		t.Fatalf("sibling scope should accept slugs that only repeat under different parents: %v", ps)
	}
	nodes = append(nodes, Node{ID: "a-y", Parents: []string{"a"}, Labels: en("y", "x")})
	if ps := Validate(nodes, WithSlugScope(SlugScopeSiblings)); ps == nil {
		t.Fatal("sibling scope should reject the repeated slug among siblings")
	}
}

func TestBuildReturnsFirstProblem(t *testing.T) {
	_, err := Build([]Node{{ID: "a", Labels: Labels{"en": {Name: "a"}}}, {ID: "b", Parents: []string{"nope"}, Labels: Labels{"en": {Name: "b"}}}})
	var p Problem
	if !errors.As(err, &p) || p.Code != ProblemParentMissing {
		t.Fatalf("want ProblemParentMissing, got %v", err)
	}
}

func TestProblemsAreSorted(t *testing.T) {
	ps := Validate([]Node{
		{ID: "z", Parents: []string{"missing"}, Labels: Labels{"en": {Name: "z"}}},
		{ID: "a", Parents: []string{"missing"}, Labels: Labels{"en": {Name: "a"}}},
		{ID: "r", Labels: Labels{"en": {Name: "r"}}},
	})
	if len(ps) != 2 || ps[0].NodeID != "a" || ps[1].NodeID != "z" {
		t.Fatalf("problems should be sorted by node id: %v", ps)
	}
	if ps.Error() == "" {
		t.Fatal("Problems.Error should describe the findings")
	}
}

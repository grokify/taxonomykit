package taxonomykit

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Problem codes reported by Validate.
const (
	ProblemIDInvalid           = "id.invalid"
	ProblemIDDuplicate         = "id.duplicate"
	ProblemParentMissing       = "parent.missing"
	ProblemParentSelf          = "parent.self"
	ProblemParentDuplicate     = "parent.duplicate"
	ProblemCycle               = "cycle"
	ProblemRootNone            = "root.none"
	ProblemLabelDefaultMissing = "label.default.missing"
	ProblemSlugDuplicate       = "slug.duplicate"
	ProblemLangInvalid         = "lang.invalid"
)

// Problem is one validation finding.
type Problem struct {
	Code    string
	NodeID  string
	Lang    string
	Message string
}

func (p Problem) Error() string {
	var b strings.Builder
	b.WriteString(p.Code)
	if p.NodeID != "" {
		b.WriteString(" [")
		b.WriteString(p.NodeID)
		if p.Lang != "" {
			b.WriteString("@")
			b.WriteString(p.Lang)
		}
		b.WriteString("]")
	}
	if p.Message != "" {
		b.WriteString(": ")
		b.WriteString(p.Message)
	}
	return b.String()
}

// Problems is every finding from one Validate call. It is an error so callers
// can return it directly.
type Problems []Problem

func (ps Problems) Error() string {
	switch len(ps) {
	case 0:
		return "no problems"
	case 1:
		return ps[0].Error()
	}
	return fmt.Sprintf("%s (and %d more)", ps[0].Error(), len(ps)-1)
}

var (
	idPattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)
	langPattern = regexp.MustCompile(`^[a-z]{2,3}(-[a-z0-9]{1,8})*$`)
)

// SlugScope controls where a slug must be unique within one language.
type SlugScope int

const (
	// SlugScopeTree requires slugs to be unique across the whole tree, as
	// flat "/<slug>/" URL schemes need.
	SlugScopeTree SlugScope = iota
	// SlugScopeSiblings requires uniqueness only among nodes sharing a
	// parent, as path-style "/<a>/<b>/" URL schemes need.
	SlugScopeSiblings
)

// Validate checks nodes against every structural rule and returns all
// findings, ordered by node ID then code, so a tool can report them at once.
// Language tags are compared after NormalizeLang. A nil result means the
// nodes build a valid tree.
func Validate(nodes []Node, opts ...Option) Problems {
	o := applyOptions(opts)
	var ps Problems
	add := func(code, id, lang, msg string) {
		ps = append(ps, Problem{Code: code, NodeID: id, Lang: lang, Message: msg})
	}

	byID := make(map[string]*Node, len(nodes))
	valid := make([]bool, len(nodes)) // first occurrence of a well-formed id
	for i := range nodes {
		n := &nodes[i]
		if !idPattern.MatchString(n.ID) {
			add(ProblemIDInvalid, n.ID, "", "id must match "+idPattern.String())
			continue
		}
		if _, dup := byID[n.ID]; dup {
			add(ProblemIDDuplicate, n.ID, "", "id appears more than once")
			continue
		}
		byID[n.ID] = n
		valid[i] = true
	}

	roots := 0
	for i, n := range nodes {
		if !valid[i] {
			continue // invalid or duplicate id; already reported
		}
		if len(n.Parents) == 0 {
			roots++
		}
		seen := map[string]bool{}
		for _, p := range n.Parents {
			switch {
			case p == n.ID:
				add(ProblemParentSelf, n.ID, "", "node lists itself as a parent")
			case seen[p]:
				add(ProblemParentDuplicate, n.ID, "", "parent "+p+" listed more than once")
			default:
				if _, ok := byID[p]; !ok {
					add(ProblemParentMissing, n.ID, "", "parent "+p+" does not exist")
				}
			}
			seen[p] = true
		}
		if _, ok := labelIn(n.Labels, o.defaultLang); !ok {
			add(ProblemLabelDefaultMissing, n.ID, o.defaultLang, "no name in the default language")
		}
		for lang := range n.Labels {
			if !langPattern.MatchString(NormalizeLang(lang)) {
				add(ProblemLangInvalid, n.ID, lang, "language tag is not a BCP-47 tag")
			}
		}
	}
	if len(nodes) > 0 && roots == 0 {
		add(ProblemRootNone, "", "", "no node without parents")
	}

	ps = append(ps, findCycles(byID)...)
	ps = append(ps, findSlugDuplicates(nodes, valid, o.slugScope)...)

	sort.SliceStable(ps, func(i, j int) bool {
		if ps[i].NodeID != ps[j].NodeID {
			return ps[i].NodeID < ps[j].NodeID
		}
		if ps[i].Code != ps[j].Code {
			return ps[i].Code < ps[j].Code
		}
		return ps[i].Lang < ps[j].Lang
	})
	if len(ps) == 0 {
		return nil
	}
	return ps
}

// labelIn returns the label stored under lang, matching after normalization.
func labelIn(labels Labels, lang string) (Label, bool) {
	want := NormalizeLang(lang)
	for k, l := range labels {
		if NormalizeLang(k) == want && l.Name != "" {
			return l, true
		}
	}
	return Label{}, false
}

// findCycles reports one problem per node that lies on a parent cycle.
func findCycles(byID map[string]*Node) Problems {
	const (
		white = iota
		grey
		black
	)
	color := make(map[string]int, len(byID))
	onCycle := map[string]bool{}
	var visit func(id string, stack []string)
	visit = func(id string, stack []string) {
		color[id] = grey
		stack = append(stack, id)
		for _, p := range byID[id].Parents {
			if _, ok := byID[p]; !ok {
				continue
			}
			switch color[p] {
			case white:
				visit(p, stack)
			case grey:
				for i := len(stack) - 1; i >= 0; i-- {
					onCycle[stack[i]] = true
					if stack[i] == p {
						break
					}
				}
			}
		}
		color[id] = black
	}
	ids := sortedIDs(byID)
	for _, id := range ids {
		if color[id] == white {
			visit(id, nil)
		}
	}
	var ps Problems
	for _, id := range ids {
		if onCycle[id] {
			ps = append(ps, Problem{Code: ProblemCycle, NodeID: id, Message: "node is on a parent cycle"})
		}
	}
	return ps
}

// findSlugDuplicates reports a problem for every node whose slug collides
// with another node's slug in the same language within the scope.
func findSlugDuplicates(nodes []Node, valid []bool, scope SlugScope) Problems {
	type key struct{ lang, scope, slug string }
	owners := map[key][]string{}
	for i, n := range nodes {
		if !valid[i] {
			continue
		}
		for lang, l := range n.Labels {
			if l.Slug == "" {
				continue
			}
			scopes := []string{""}
			if scope == SlugScopeSiblings {
				scopes = n.Parents
				if len(scopes) == 0 {
					scopes = []string{""}
				}
			}
			for _, s := range scopes {
				k := key{NormalizeLang(lang), s, l.Slug}
				owners[k] = append(owners[k], n.ID)
			}
		}
	}
	var ps Problems
	seen := map[string]bool{}
	for k, ids := range owners {
		if len(ids) < 2 {
			continue
		}
		sort.Strings(ids)
		for _, id := range ids {
			sk := id + "\x00" + k.lang + "\x00" + k.slug
			if seen[sk] {
				continue
			}
			seen[sk] = true
			ps = append(ps, Problem{Code: ProblemSlugDuplicate, NodeID: id, Lang: k.lang,
				Message: fmt.Sprintf("slug %q is also used by %s", k.slug, strings.Join(others(ids, id), ", "))})
		}
	}
	return ps
}

func others(ids []string, self string) []string {
	out := make([]string, 0, len(ids)-1)
	for _, id := range ids {
		if id != self {
			out = append(out, id)
		}
	}
	return out
}

func sortedIDs(byID map[string]*Node) []string {
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

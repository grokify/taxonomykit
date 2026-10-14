package taxonomykit

// fixture mirrors the shape of a real topic tree: two languages, a node
// with two parents, explicit ordering on some roots' children, a root.
func fixture() []Node {
	en := func(name, slug string) Labels { return Labels{"en": {Name: name, Slug: slug}} }
	two := func(name, slug, esName, esSlug string) Labels {
		return Labels{"en": {Name: name, Slug: slug}, "es": {Name: esName, Slug: esSlug}}
	}
	order := func(i int) *int { return &i }
	return []Node{
		{ID: "all", Labels: two("All", "all", "Todo", "todo")},
		{ID: "business", Parents: []string{"all"}, Labels: en("Business", "business")},
		{ID: "law", Parents: []string{"all"}, Labels: two("Law", "law", "Derecho", "derecho")},
		{ID: "science", Parents: []string{"all"}, Labels: en("Science", "science"), Order: order(1)},
		{ID: "technology", Parents: []string{"all"}, Labels: en("Technology", "technology"), Order: order(0)},
		{ID: "finance", Parents: []string{"business"}, Labels: en("Finance", "finance")},
		{ID: "realestate", Parents: []string{"business"}, Labels: en("Real Estate", "realestate")},
		{ID: "personal-finance", Parents: []string{"finance"}, Labels: en("Personal Finance", "personal-finance")},
		{ID: "mortgage", Parents: []string{"personal-finance", "realestate"}, Labels: en("Mortgage", "mortgage")},
		{ID: "patent", Parents: []string{"law"}, Labels: en("Patent", "patent")},
		{ID: "trademark", Parents: []string{"law"}, Labels: en("Trademark", "trademark")},
		{ID: "water", Parents: []string{"science"}, Labels: en("Water", "water")},
	}
}

func ids(ns []*Node) []string {
	out := make([]string, len(ns))
	for i, n := range ns {
		out[i] = n.ID
	}
	return out
}

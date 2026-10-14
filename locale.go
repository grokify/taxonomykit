package taxonomykit

import "strings"

// NormalizeLang lowercases and trims a BCP-47 tag so "en-US", " EN_us " and
// "en-us" compare equal.
func NormalizeLang(lang string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(lang), "_", "-"))
}

// FallbackChain returns the languages tried for a request, most specific
// first: the tag itself, each shorter prefix ("pt-br-x" → "pt-br" → "pt"),
// then the default language.
func FallbackChain(lang, defaultLang string) []string {
	lang = NormalizeLang(lang)
	defaultLang = NormalizeLang(defaultLang)
	var chain []string
	for lang != "" {
		chain = append(chain, lang)
		i := strings.LastIndex(lang, "-")
		if i < 0 {
			break
		}
		lang = lang[:i]
	}
	for _, c := range chain {
		if c == defaultLang {
			return chain
		}
	}
	return append(chain, defaultLang)
}

// Label returns the node's label for lang after fallback, the language the
// label was actually taken from, and whether the node exists and has any
// usable label. A label is usable when its Name is non-empty.
func (t *Tree) Label(id, lang string) (Label, string, bool) {
	n, ok := t.byID[id]
	if !ok {
		return Label{}, "", false
	}
	for _, l := range FallbackChain(lang, t.opts.defaultLang) {
		if lab, ok := n.Labels[l]; ok && lab.Name != "" {
			return lab, l, true
		}
	}
	return Label{}, "", false
}

// Name is the node's name for lang after fallback, or "" if unknown.
func (t *Tree) Name(id, lang string) string {
	l, _, _ := t.Label(id, lang)
	return l.Name
}

// Slug is the node's URL slug for lang after fallback, or "" if the node has
// no slug in any language on the chain.
func (t *Tree) Slug(id, lang string) string {
	n, ok := t.byID[id]
	if !ok {
		return ""
	}
	for _, l := range FallbackChain(lang, t.opts.defaultLang) {
		if lab, ok := n.Labels[l]; ok && lab.Slug != "" {
			return lab.Slug
		}
	}
	return ""
}

// Breadcrumb is the primary lineage resolved for lang, root first. Each
// Crumb records the language its Name and Slug came from.
func (t *Tree) Breadcrumb(id, lang string) []Crumb {
	lin := t.Lineage(id)
	if lin == nil {
		return nil
	}
	out := make([]Crumb, len(lin))
	for i, n := range lin {
		l, used, _ := t.Label(n.ID, lang)
		out[i] = Crumb{ID: n.ID, Name: l.Name, Slug: t.Slug(n.ID, lang), Lang: used}
	}
	return out
}

// BySlug finds the node whose effective slug in lang (after fallback) is
// slug. Explicit slugs in lang take precedence over fallback slugs. The
// language must be one that appears in the tree or the default language.
func (t *Tree) BySlug(lang, slug string) (*Node, bool) {
	lang = NormalizeLang(lang)
	idx, ok := t.slugs[lang]
	if !ok {
		// Unknown language: resolve through the chain to a known one.
		for _, l := range FallbackChain(lang, t.opts.defaultLang) {
			if idx, ok = t.slugs[l]; ok {
				break
			}
		}
		if !ok {
			return nil, false
		}
	}
	n, ok := idx[slug]
	return n, ok
}

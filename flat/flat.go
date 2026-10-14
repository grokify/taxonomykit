// Package flat reads and writes taxonomies as delimited text (CSV, TSV, PSV)
// with one row per node and language.
//
// A file has a header row naming its columns, or the caller supplies a
// Columns mapping for headerless files. Recognized column names, matched
// case-insensitively: lang, parent, id, slug, name, alt, order. Any other
// column is stored in Node.Attrs under its header name. Lines that are blank
// or start with '#' are skipped.
//
// Several rows may describe one node: one per language, or repeated rows
// with different parents. The first row seen for an id defines Parents[0];
// a later row naming a different, non-empty parent adds that parent, so a
// legacy single-parent file that listed a node twice becomes a multi-parent
// document rather than silently keeping the last parent.
package flat

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/grokify/taxonomykit"
)

// Format is the delimiter convention of a file.
type Format int

const (
	CSV Format = iota // comma-separated, RFC 4180 quoting
	TSV               // tab-separated, no quoting
	PSV               // pipe-separated, cells trimmed of surrounding whitespace
)

// String returns the lowercase format name, which is also its file extension.
func (f Format) String() string {
	switch f {
	case CSV:
		return "csv"
	case TSV:
		return "tsv"
	case PSV:
		return "psv"
	}
	return "unknown"
}

// ParseFormat maps "csv", "tsv" or "psv" (any case, optional leading dot) to
// a Format.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimPrefix(s, ".")) {
	case "csv":
		return CSV, nil
	case "tsv":
		return TSV, nil
	case "psv":
		return PSV, nil
	}
	return CSV, fmt.Errorf("flat: unknown format %q", s)
}

// Columns maps roles to zero-based column positions for headerless files.
// Positions that do not apply are -1.
type Columns struct {
	Lang, Parent, ID, Slug, Name, Alt, Order int
	// Attrs maps extra column positions to attribute names.
	Attrs map[int]string
}

// Options configures Read.
type Options struct {
	// Columns is used when the file has no header row.
	Columns *Columns
	// DefaultLang becomes Document.DefaultLang. When empty, the language of
	// the first data row is used, or "en" if the file has no lang column.
	DefaultLang string
	// AltSeparator splits the alt column into alternative names; default ";".
	AltSeparator string
}

// Read parses delimited text into a Document. Nodes keep first-appearance
// order; rows are merged per node as described in the package comment.
func Read(r io.Reader, f Format, opts Options) (*taxonomykit.Document, error) {
	rows, err := readRows(r, f)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &taxonomykit.Document{Version: taxonomykit.DocumentVersion, DefaultLang: pick(opts.DefaultLang, "en")}, nil
	}
	cols := opts.Columns
	if cols == nil {
		c, err := columnsFromHeader(rows[0])
		if err != nil {
			return nil, err
		}
		cols = c
		rows = rows[1:]
	}
	if cols.ID < 0 || cols.Name < 0 {
		return nil, errors.New("flat: id and name columns are required")
	}
	altSep := pick(opts.AltSeparator, ";")

	doc := &taxonomykit.Document{Version: taxonomykit.DocumentVersion, DefaultLang: taxonomykit.NormalizeLang(opts.DefaultLang)}
	index := map[string]int{}
	for i, row := range rows {
		get := func(pos int) string {
			if pos < 0 || pos >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[pos])
		}
		id := get(cols.ID)
		if id == "" {
			return nil, fmt.Errorf("flat: row %d has an empty id", i+1)
		}
		lang := taxonomykit.NormalizeLang(get(cols.Lang))
		if lang == "" {
			lang = pick(doc.DefaultLang, "en")
		}
		if doc.DefaultLang == "" {
			doc.DefaultLang = lang
		}
		pos, seen := index[id]
		if !seen {
			pos = len(doc.Nodes)
			index[id] = pos
			doc.Nodes = append(doc.Nodes, taxonomykit.Node{ID: id, Labels: taxonomykit.Labels{}})
		}
		n := &doc.Nodes[pos]
		if parent := get(cols.Parent); parent != "" && parent != id && !contains(n.Parents, parent) {
			n.Parents = append(n.Parents, parent)
		}
		label := n.Labels[lang]
		if name := get(cols.Name); name != "" {
			label.Name = name
		}
		if slug := get(cols.Slug); slug != "" {
			label.Slug = slug
		}
		if alt := get(cols.Alt); alt != "" {
			for _, a := range strings.Split(alt, altSep) {
				if a = strings.TrimSpace(a); a != "" && !contains(label.Alt, a) {
					label.Alt = append(label.Alt, a)
				}
			}
		}
		n.Labels[lang] = label
		if o := get(cols.Order); o != "" && n.Order == nil {
			v, err := strconv.Atoi(o)
			if err != nil {
				return nil, fmt.Errorf("flat: row %d: order %q is not an integer", i+1, o)
			}
			n.Order = &v
		}
		for p, name := range cols.Attrs {
			if v := get(p); v != "" {
				if n.Attrs == nil {
					n.Attrs = map[string]string{}
				}
				if _, set := n.Attrs[name]; !set {
					n.Attrs[name] = v
				}
			}
		}
	}
	return doc, nil
}

// Write emits a header row and one row per node and language, nodes in
// document order and languages with the default first then sorted. Attrs
// keys become extra columns, sorted.
func Write(w io.Writer, doc *taxonomykit.Document, f Format) error {
	attrKeys := map[string]bool{}
	for _, n := range doc.Nodes {
		for k := range n.Attrs {
			attrKeys[k] = true
		}
	}
	extra := make([]string, 0, len(attrKeys))
	for k := range attrKeys {
		extra = append(extra, k)
	}
	sort.Strings(extra)

	header := append([]string{"lang", "parent", "id", "slug", "name", "alt", "order"}, extra...)
	rows := [][]string{header}
	def := taxonomykit.NormalizeLang(doc.DefaultLang)
	for _, n := range doc.Nodes {
		langs := make([]string, 0, len(n.Labels))
		for l := range n.Labels {
			langs = append(langs, taxonomykit.NormalizeLang(l))
		}
		sort.Slice(langs, func(i, j int) bool {
			if (langs[i] == def) != (langs[j] == def) {
				return langs[i] == def
			}
			return langs[i] < langs[j]
		})
		parents := n.Parents
		if len(parents) == 0 {
			parents = []string{""}
		}
		for li, lang := range langs {
			label := n.Labels[lang]
			// The first language row carries every parent (one row per
			// parent); later language rows repeat only the primary parent.
			ps := parents[:1]
			if li == 0 {
				ps = parents
			}
			for _, p := range ps {
				row := []string{lang, p, n.ID, label.Slug, label.Name, strings.Join(label.Alt, ";"), ""}
				if n.Order != nil {
					row[6] = strconv.Itoa(*n.Order)
				}
				for _, k := range extra {
					row = append(row, n.Attrs[k])
				}
				rows = append(rows, row)
			}
		}
	}
	return writeRows(w, f, rows)
}

func readRows(r io.Reader, f Format) ([][]string, error) {
	if f == CSV {
		cr := csv.NewReader(r)
		cr.Comment = '#'
		cr.FieldsPerRecord = -1
		cr.TrimLeadingSpace = true
		recs, err := cr.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("flat: %w", err)
		}
		var out [][]string
		for _, rec := range recs {
			if !blank(rec) {
				out = append(out, rec)
			}
		}
		return out, nil
	}
	sep := "\t"
	if f == PSV {
		sep = "|"
	}
	var out [][]string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if t := strings.TrimSpace(line); t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		cells := strings.Split(line, sep)
		if f == PSV {
			for i := range cells {
				cells[i] = strings.TrimSpace(cells[i])
			}
		}
		out = append(out, cells)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("flat: %w", err)
	}
	return out, nil
}

func writeRows(w io.Writer, f Format, rows [][]string) error {
	if f == CSV {
		cw := csv.NewWriter(w)
		if err := cw.WriteAll(rows); err != nil {
			return fmt.Errorf("flat: %w", err)
		}
		return nil
	}
	sep, forbidden := "\t", "\t\r\n"
	if f == PSV {
		sep, forbidden = " | ", "|\r\n"
	}
	bw := bufio.NewWriter(w)
	for _, row := range rows {
		for _, c := range row {
			if strings.ContainsAny(c, forbidden) {
				return fmt.Errorf("flat: cell %q contains the %s delimiter or a newline", c, f)
			}
		}
		if _, err := bw.WriteString(strings.Join(row, sep) + "\n"); err != nil {
			return fmt.Errorf("flat: %w", err)
		}
	}
	return bw.Flush()
}

func columnsFromHeader(header []string) (*Columns, error) {
	c := &Columns{Lang: -1, Parent: -1, ID: -1, Slug: -1, Name: -1, Alt: -1, Order: -1, Attrs: map[int]string{}}
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "lang", "language":
			c.Lang = i
		case "parent", "parentid", "parent_id":
			c.Parent = i
		case "id":
			c.ID = i
		case "slug":
			c.Slug = i
		case "name", "display", "label":
			c.Name = i
		case "alt":
			c.Alt = i
		case "order":
			c.Order = i
		case "":
			return nil, fmt.Errorf("flat: header column %d is empty", i+1)
		default:
			c.Attrs[i] = strings.TrimSpace(h)
		}
	}
	return c, nil
}

func blank(rec []string) bool {
	for _, c := range rec {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

func pick(v, def string) string {
	if v != "" {
		return v
	}
	return def
}

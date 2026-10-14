package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/grokify/taxonomykit"
	"github.com/spf13/cobra"
)

func newValidate() *cobra.Command {
	var in inputFlags
	cmd := &cobra.Command{
		Use:   "validate <file>",
		Short: "Report every structural problem in a taxonomy file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := in.load(args[0])
			if err != nil {
				return err
			}
			ps := doc.Validate()
			for _, p := range ps {
				fmt.Fprintln(cmd.OutOrStdout(), p.Error())
			}
			if len(ps) > 0 {
				return fmt.Errorf("%d problem(s)", len(ps))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ok: %d nodes, languages %s\n", len(doc.Nodes), strings.Join(langsOf(doc), ","))
			return nil
		},
	}
	in.bind(cmd)
	return cmd
}

func newConvert() *cobra.Command {
	var in inputFlags
	var to, out string
	cmd := &cobra.Command{
		Use:   "convert <file>",
		Short: "Convert a taxonomy between csv, tsv, psv and json",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := in.load(args[0])
			if err != nil {
				return err
			}
			if ps := doc.Validate(); ps != nil {
				return ps
			}
			w := cmd.OutOrStdout()
			if out != "" {
				fh, err := os.Create(out)
				if err != nil {
					return err
				}
				defer fh.Close()
				w = fh
			}
			return write(w, doc, to)
		},
	}
	in.bind(cmd)
	cmd.Flags().StringVar(&to, "to", "json", "output format: csv, tsv, psv or json")
	cmd.Flags().StringVarP(&out, "output", "o", "", "output file (default: stdout)")
	return cmd
}

func newTree() *cobra.Command {
	var in inputFlags
	var lang, root string
	var depth int
	cmd := &cobra.Command{
		Use:   "tree <file>",
		Short: "Print the taxonomy as an indented tree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tr, err := loadTree(&in, args[0])
			if err != nil {
				return err
			}
			if lang == "" {
				lang = tr.DefaultLang()
			}
			print := func(n *taxonomykit.Node, d int) bool {
				if depth >= 0 && d > depth {
					return false
				}
				slug := tr.Slug(n.ID, lang)
				fmt.Fprintf(cmd.OutOrStdout(), "%s%s (%s) /%s\n", strings.Repeat("  ", d), tr.Name(n.ID, lang), n.ID, slug)
				return true
			}
			if root == "" {
				tr.Walk(print)
				return nil
			}
			if !tr.Has(root) {
				return fmt.Errorf("unknown node %q", root)
			}
			tr.WalkFrom(root, print)
			return nil
		},
	}
	in.bind(cmd)
	cmd.Flags().StringVar(&lang, "lang", "", "language for names and slugs (default: the document's default language)")
	cmd.Flags().StringVar(&root, "root", "", "print only the subtree below this node id")
	cmd.Flags().IntVar(&depth, "depth", -1, "maximum depth to print (-1 for all)")
	return cmd
}

func newLineage() *cobra.Command {
	var in inputFlags
	var lang string
	var all bool
	cmd := &cobra.Command{
		Use:   "lineage <file> <id>",
		Short: "Print the path from the root to a node",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			tr, err := loadTree(&in, args[0])
			if err != nil {
				return err
			}
			if !tr.Has(args[1]) {
				return fmt.Errorf("unknown node %q", args[1])
			}
			if lang == "" {
				lang = tr.DefaultLang()
			}
			lineages := [][]*taxonomykit.Node{tr.Lineage(args[1])}
			if all {
				lineages = tr.Lineages(args[1])
			}
			for _, lin := range lineages {
				parts := make([]string, len(lin))
				for i, n := range lin {
					parts[i] = fmt.Sprintf("%s (%s)", tr.Name(n.ID, lang), n.ID)
				}
				fmt.Fprintln(cmd.OutOrStdout(), strings.Join(parts, " > "))
			}
			return nil
		},
	}
	in.bind(cmd)
	cmd.Flags().StringVar(&lang, "lang", "", "language for names (default: the document's default language)")
	cmd.Flags().BoolVar(&all, "all", false, "print every lineage of a multi-parent node, primary first")
	return cmd
}

func loadTree(in *inputFlags, path string) (*taxonomykit.Tree, error) {
	doc, err := in.load(path)
	if err != nil {
		return nil, err
	}
	tr, err := doc.Build()
	if err != nil {
		var ps taxonomykit.Problems
		if errors.As(err, &ps) {
			return nil, ps
		}
		return nil, err
	}
	return tr, nil
}

func langsOf(doc *taxonomykit.Document) []string {
	seen := map[string]bool{}
	var out []string
	for _, n := range doc.Nodes {
		for l := range n.Labels {
			l = taxonomykit.NormalizeLang(l)
			if !seen[l] {
				seen[l] = true
				out = append(out, l)
			}
		}
	}
	sort.Strings(out)
	return out
}

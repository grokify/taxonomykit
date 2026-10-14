package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/grokify/taxonomykit"
	"github.com/grokify/taxonomykit/flat"
	"github.com/spf13/cobra"
)

// inputFlags are shared by every command that reads a taxonomy file.
type inputFlags struct {
	from        string
	defaultLang string
}

func (f *inputFlags) bind(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.from, "from", "", "input format: csv, tsv, psv or json (default: from the file extension)")
	cmd.Flags().StringVar(&f.defaultLang, "default-lang", "", "default language for flat files (default: the first row's language)")
}

// load reads a document from path, or from stdin when path is "-".
func (f *inputFlags) load(path string) (*taxonomykit.Document, error) {
	format := f.from
	if format == "" {
		format = strings.TrimPrefix(filepath.Ext(path), ".")
	}
	if format == "" {
		return nil, fmt.Errorf("cannot infer the format of %q; pass --from", path)
	}
	var r io.Reader = os.Stdin
	if path != "-" {
		fh, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer fh.Close()
		r = fh
	}
	if strings.EqualFold(format, "json") {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return taxonomykit.ParseDocument(data)
	}
	ff, err := flat.ParseFormat(format)
	if err != nil {
		return nil, err
	}
	return flat.Read(r, ff, flat.Options{DefaultLang: f.defaultLang})
}

// write encodes doc in the named format to w.
func write(w io.Writer, doc *taxonomykit.Document, format string) error {
	if strings.EqualFold(format, "json") {
		data, err := doc.JSON()
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	ff, err := flat.ParseFormat(format)
	if err != nil {
		return err
	}
	return flat.Write(w, doc, ff)
}

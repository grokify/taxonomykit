// Command taxonomykit validates, converts and prints taxonomy files.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "taxonomykit:", err)
		os.Exit(1)
	}
}

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "taxonomykit",
		Short:         "Validate, convert and print hierarchical taxonomies",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newValidate(), newConvert(), newTree(), newLineage())
	return root
}

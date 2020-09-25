package clihelpers

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ory/x/cmdx"
)

type (
	OutputHeader interface {
		Header() []string
	}
	OutputEntry interface {
		OutputHeader
		Fields() []string
		Interface() interface{}
	}
	OutputCollection interface {
		OutputHeader
		Table() [][]string
		Interface() interface{}
		Len() int
	}

	format string
)

const (
	FormatQuiet      format = "quiet"
	FormatTable      format = "table"
	FormatJSON       format = "json"
	FormatJSONPretty format = "json-pretty"

	FlagQuiet  = "quiet"
	FlagFormat = "format"

	None = "<none>"
)

func PrintErrors(cmd *cobra.Command, errs map[string]error) {
	for src, err := range errs {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s: %s\n", src, err.Error())
	}
}

func PrintRow(cmd *cobra.Command, row OutputEntry) {
	f := getFormat(cmd)

	switch f {
	case FormatQuiet:
		fmt.Fprintln(cmd.OutOrStdout(), row.Fields()[0])
	case FormatJSON:
		printJSON(cmd.OutOrStdout(), row.Interface(), false)
	case FormatJSONPretty:
		printJSON(cmd.OutOrStdout(), row.Interface(), true)
	case FormatTable:
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 1, '\t', 0)

		fields := row.Fields()
		for i, h := range row.Header() {
			fmt.Fprintf(w, "%s\t%s\t\n", h, fields[i])
		}

		w.Flush()
	}
}

func PrintCollection(cmd *cobra.Command, collection OutputCollection) {
	if collection.Len() == 0 {
		// don't print headers, ... when there is no content
		return
	}
	f := getFormat(cmd)

	switch f {
	case FormatQuiet:
		for _, row := range collection.Table() {
			fmt.Fprintln(cmd.OutOrStdout(), row[0])
		}
	case FormatJSON:
		printJSON(cmd.OutOrStdout(), collection.Interface(), false)
	case FormatJSONPretty:
		printJSON(cmd.OutOrStdout(), collection.Interface(), true)
	case FormatTable:
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 1, '\t', 0)

		for _, h := range collection.Header() {
			fmt.Fprintf(w, "%s\t", h)
		}
		fmt.Fprintln(w)

		for _, row := range collection.Table() {
			fmt.Fprintln(w, strings.Join(row, "\t")+"\t")
		}

		w.Flush()
	}
}

func getFormat(cmd *cobra.Command) format {
	q, err := cmd.Flags().GetBool(FlagQuiet)
	// unexpected error
	cmdx.Must(err, "flag access error: %s", err)

	if q {
		return FormatQuiet
	}

	f, err := cmd.Flags().GetString(FlagFormat)
	// unexpected error
	cmdx.Must(err, "flag access error: %s", err)

	switch f {
	case string(FormatTable):
		return FormatTable
	case string(FormatJSON):
		return FormatJSON
	case string(FormatJSONPretty):
		return FormatJSONPretty
	}

	// default format is table
	return FormatTable
}

func printJSON(w io.Writer, v interface{}, pretty bool) {
	e := json.NewEncoder(w)
	if pretty {
		e.SetIndent("", "  ")
	}
	err := e.Encode(v)
	// unexpected error
	cmdx.Must(err, "Error encoding JSON: %s", err)
}

func RegisterFormatFlags(flags *pflag.FlagSet) {
	flags.BoolP(FlagQuiet, FlagQuiet[:1], false, "Prints only IDs, one per line. Takes precedence over --format.")
	flags.StringP(FlagFormat, FlagFormat[:1], "", fmt.Sprintf("Set the output format. One of %s, %s, and %s.", FormatTable, FormatJSON, FormatJSONPretty))
}

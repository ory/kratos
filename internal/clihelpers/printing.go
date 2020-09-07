package clihelpers

import (
	"encoding/json"
	"fmt"
	"os"
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
	}

	format string
)

const (
	formatQuiet      format = "quiet"
	formatTable      format = "table"
	formatJSON       format = "json"
	formatJSONPretty format = "json-pretty"

	flagQuiet  = "quiet"
	flagFormat = "format"

	None = "<none>"
)

func PrintErrors(_ *cobra.Command, errs map[string]error) {
	for src, err := range errs {
		fmt.Fprintf(os.Stderr, "%s: %s\n", src, err.Error())
	}
}

func PrintRow(cmd *cobra.Command, row OutputEntry) {
	f := getFormat(cmd)

	switch f {
	case formatQuiet:
		fmt.Println(row.Fields()[0])
	case formatJSON:
		printJSON(row.Interface(), false)
	case formatJSONPretty:
		printJSON(row.Interface(), true)
	case formatTable:
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)

		fields := row.Fields()
		for i, h := range row.Header() {
			fmt.Fprintf(w, "%s\t%s\t\n", h, fields[i])
		}

		w.Flush()
	}
}

func PrintCollection(cmd *cobra.Command, collection OutputCollection) {
	f := getFormat(cmd)

	switch f {
	case formatQuiet:
		for _, row := range collection.Table() {
			fmt.Println(row[0])
		}
	case formatJSON:
		printJSON(collection.Interface(), false)
	case formatJSONPretty:
		printJSON(collection.Interface(), true)
	case formatTable:
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)

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
	q, err := cmd.Flags().GetBool(flagQuiet)
	cmdx.Must(err, "flag access error: %s", err)

	if q {
		return formatQuiet
	}

	f, err := cmd.Flags().GetString(flagFormat)
	cmdx.Must(err, "flag access error: %s", err)

	switch f {
	case string(formatTable):
		return formatTable
	case string(formatJSON):
		return formatJSON
	case string(formatJSONPretty):
		return formatJSONPretty
	}

	// output is a terminal, default is table
	if info, _ := os.Stdout.Stat(); info.Mode()&os.ModeCharDevice != 0 {
		return formatTable
	}

	// output is not a terminal, default is quiet
	return formatQuiet
}

func printJSON(v interface{}, pretty bool) {
	e := json.NewEncoder(os.Stdout)
	if pretty {
		e.SetIndent("", "  ")
	}
	err := e.Encode(v)
	cmdx.Must(err, "Error encoding JSON: %s", err)
}

func RegisterFormatFlags(flags *pflag.FlagSet) {
	flags.BoolP(flagQuiet, flagQuiet[:1], false, "Prints only IDs, one per line. Takes precedence over --format.")
	flags.StringP(flagFormat, flagFormat[:1], "", fmt.Sprintf("Set the output format. One of %s, %s, and %s.", formatTable, formatJSON, formatJSONPretty))
}

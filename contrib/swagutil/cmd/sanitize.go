package cmd

import (
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/x/cmdx"
)

func sanitizeIter(raw string) string {
	result := raw
	gjson.Parse(raw).ForEach(func(key, value gjson.Result) bool {
		var err error
		if !key.Exists() {
			return true
		}

		switch value.Type {
		case gjson.JSON:
			result, err = sjson.SetRaw(result, key.String(), sanitizeIter(value.Raw))
			cmdx.Must(err, "could not update path (%s - %s): %s", key.Raw, value.Raw, err)
		case gjson.String:
			switch key.String() {
			case "x-go-package":
				fallthrough
			case "x-go-name":
				result, err = sjson.Delete(result, key.String())
				cmdx.Must(err, "could not delete path (%s - %s): %s", key.Raw, value.Raw, err)
			}
		}
		return true
	})
	return result
}

// sanitizeCmd represents the sanitize command
var sanitizeCmd = &cobra.Command{
	Use:  "sanitize <file>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file, err := ioutil.ReadFile(args[0])
		cmdx.Must(err, "Unable to read file: %s", err)

		result := []byte(sanitizeIter(string(file)))
		result, err = sjson.SetRawBytes(result, "definitions.UUID", []byte(`{"type": "string", "format": "uuid4"}`))
		cmdx.Must(err, "could not set definitions.UUID: %s", err)

		err = os.Remove(args[0])
		cmdx.Must(err, "Unable to delete file: %s", err)

		err = ioutil.WriteFile(args[0], result, 0766)
		cmdx.Must(err, "Unable to write file: %s", err)
	},
}

func init() {
	rootCmd.AddCommand(sanitizeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sanitizeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sanitizeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

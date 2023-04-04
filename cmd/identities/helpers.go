// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identities

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/ory/x/cmdx"
)

func parseIdentities(raw []byte) (rawIdentities []string) {
	res := gjson.ParseBytes(raw)
	if !res.IsArray() {
		return []string{res.Raw}
	}
	res.ForEach(func(_, v gjson.Result) bool {
		rawIdentities = append(rawIdentities, v.Raw)
		return true
	})
	return
}

func readIdentities(cmd *cobra.Command, args []string) (map[string]string, error) {
	rawIdentities := make(map[string]string)
	if len(args) == 0 {
		fc, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "STD_IN: Could not read: %s\n", err)
			return nil, cmdx.FailSilently(cmd)
		}
		for i, id := range parseIdentities(fc) {
			rawIdentities[fmt.Sprintf("STD_IN[%d]", i)] = id
		}
		return rawIdentities, nil
	}
	for _, fn := range args {
		fc, err := os.ReadFile(fn)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s: Could not open identity file: %s\n", fn, err)
			return nil, cmdx.FailSilently(cmd)
		}
		for i, id := range parseIdentities(fc) {
			rawIdentities[fmt.Sprintf("%s[%d]", fn, i)] = id
		}
	}
	return rawIdentities, nil
}

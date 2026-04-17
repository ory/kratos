// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cliclient

import (
	"fmt"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/x"
	"github.com/ory/pop/v6"
	"github.com/ory/x/flagx"
)

// phonePattern matches values that look like phone numbers: + followed by at least one digit,
// possibly with spaces, dashes, or parens. This filters out OIDC subjects and other
// identifiers that happen to start with +.
var phonePattern = regexp.MustCompile(`^\+[\d\s\-().]+$`)

type NormalizePhoneHandler struct{}

func NewNormalizePhoneHandler() *NormalizePhoneHandler {
	return &NormalizePhoneHandler{}
}

type normalizeStats struct {
	scanned int
	updated int
	skipped int
	errors  int
}

func (h *NormalizePhoneHandler) NormalizePhoneNumbers(cmd *cobra.Command, args []string, opts ...driver.RegistryOption) error {
	d, err := getPersister(cmd, args, opts)
	if err != nil {
		return err
	}

	conn := d.Persister().GetConnection(cmd.Context())
	batchSize := flagx.MustGetInt(cmd, "batch-size")
	dryRun := flagx.MustGetBool(cmd, "dry-run")
	batchDelay := flagx.MustGetDuration(cmd, "batch-delay")

	if dryRun {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Dry run mode enabled. No changes will be written.")
	}

	tables := []tableConfig{
		{
			key:  "ci",
			name: "Credential identifiers",
			selectQuery: `
			SELECT id, identifier AS value
			FROM identity_credential_identifiers
			WHERE identifier LIKE '+%'
			  AND id > ?
			ORDER BY id ASC LIMIT ?`,
			updateQuery: `UPDATE identity_credential_identifiers SET identifier = ?, updated_at = ? WHERE id = ? AND identifier = ?`,
		},
		{
			key:  "va",
			name: "Verifiable addresses",
			selectQuery: `
			SELECT id, value
			FROM identity_verifiable_addresses
			WHERE via = 'sms'
			  AND id > ?
			  AND value LIKE '+%'
			ORDER BY id ASC LIMIT ?`,
			updateQuery: `UPDATE identity_verifiable_addresses SET value = ?, updated_at = ? WHERE id = ? AND value = ?`,
		},
		{
			key:  "ra",
			name: "Recovery addresses",
			selectQuery: `
			SELECT id, value
			FROM identity_recovery_addresses
			WHERE via = 'sms'
			  AND id > ?
			  AND value LIKE '+%'
			ORDER BY id ASC LIMIT ?`,
			updateQuery: `UPDATE identity_recovery_addresses SET value = ?, updated_at = ? WHERE id = ? AND value = ?`,
		},
	}

	startAfterMap, err := parseStartAfter(flagx.MustGetStringSlice(cmd, "start-after"))
	if err != nil {
		return err
	}

	allStats := make([]normalizeStats, len(tables))
	for i, table := range tables {
		startAfter := startAfterMap[table.key]
		if startAfter != uuid.Nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s: resuming after ID %s\n", table.name, startAfter)
		}
		stats, err := normalizeTable(conn, batchSize, batchDelay, startAfter, dryRun, cmd, table)
		if err != nil {
			return errors.Wrapf(err, "normalizing %s", table.name)
		}
		allStats[i] = stats
	}

	printSummary(cmd, tables, allStats)

	return nil
}

type tableConfig struct {
	key         string
	name        string
	selectQuery string
	updateQuery string
}

// parseStartAfter parses --start-after flags of the form "key=uuid".
func parseStartAfter(args []string) (map[string]uuid.UUID, error) {
	result := make(map[string]uuid.UUID)
	for _, arg := range args {
		key, val, ok := strings.Cut(arg, "=")
		if !ok {
			return nil, errors.Errorf("invalid --start-after format %q: expected key=uuid (e.g. credentials=<id>)", arg)
		}
		id, err := uuid.FromString(val)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid UUID in --start-after %q", arg)
		}
		result[key] = id
	}
	return result, nil
}

func printSummary(cmd *cobra.Command, tables []tableConfig, allStats []normalizeStats) {
	out := cmd.OutOrStdout()

	_, _ = fmt.Fprintln(out)

	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "\tSCANNED\tUPDATED\tSKIPPED\tERRORS")

	var total normalizeStats
	for i, table := range tables {
		s := allStats[i]
		_, _ = fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\n", table.name, s.scanned, s.updated, s.skipped, s.errors)
		total.scanned += s.scanned
		total.updated += s.updated
		total.skipped += s.skipped
		total.errors += s.errors
	}
	_, _ = fmt.Fprintf(tw, "TOTAL\t%d\t%d\t%d\t%d\n", total.scanned, total.updated, total.skipped, total.errors)
	_ = tw.Flush()
}

func normalizeTable(conn *pop.Connection, batchSize int, batchDelay time.Duration, startAfter uuid.UUID, dryRun bool, cmd *cobra.Command, table tableConfig) (normalizeStats, error) {
	var stats normalizeStats
	lastID := startAfter

	for {
		var rows []struct {
			ID    uuid.UUID `db:"id"`
			Value string    `db:"value"`
		}

		if err := conn.RawQuery(table.selectQuery, lastID, batchSize).All(&rows); err != nil {
			return stats, errors.Wrapf(err, "querying %s", table.name)
		}

		if len(rows) == 0 {
			break
		}

		for _, row := range rows {
			lastID = row.ID
			stats.scanned++

			// The SQL query pre-filters with LIKE '+%'. This Go-side regex additionally
			// rejects non-phone identifiers (e.g. OIDC subjects) that happen to start
			// with '+'. We cannot use SQL regex because SQLite does not support it.
			if !phonePattern.MatchString(row.Value) {
				stats.skipped++
				continue
			}

			normalized := x.NormalizePhoneIdentifier(row.Value)
			if normalized == row.Value {
				stats.skipped++
				continue
			}

			if dryRun {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[dry-run] %s %s: %q -> %q\n", table.name, row.ID, row.Value, normalized)
				stats.updated++
				continue
			}

			now := time.Now().UTC().Truncate(time.Second)
			if err := conn.RawQuery(table.updateQuery, normalized, now, row.ID, row.Value).Exec(); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "error updating %s %s (%q -> %q): %v\n", table.name, row.ID, row.Value, normalized, err)
				stats.errors++
				continue
			}
			stats.updated++
		}

		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s\tprocessed %d rows so far (--start-after %s=%s)\n", table.name, stats.scanned, table.key, lastID)

		if batchDelay > 0 {
			time.Sleep(batchDelay)
		}
	}

	return stats, nil
}

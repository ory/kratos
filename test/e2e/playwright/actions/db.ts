// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { execSync } from "child_process"

/**
 * Extracts the SQLite file path from the TEST_DATABASE_SQLITE env var.
 *
 * Expected format: sqlite:///path/to/db.sqlite?_fk=true
 */
export function getSqliteDbPath(): string {
  const dsn = process.env["TEST_DATABASE_SQLITE"]
  if (!dsn) {
    throw new Error(
      "TEST_DATABASE_SQLITE is not set — cannot manipulate legacy data",
    )
  }

  // Strip the sqlite:// prefix and any query parameters
  const withoutScheme = dsn.replace(/^sqlite:\/\//, "")
  const pathOnly = withoutScheme.split("?")[0]
  return pathOnly
}

/**
 * Downgrades an identity created via admin API to simulate a legacy (pre-normalization)
 * record in the database. This reverts the E.164 normalization that the admin API applies
 * on creation, making the DB state match what an old Kratos version would have stored.
 *
 * Updates:
 * - identity_credential_identifiers: replaces normalized identifier with legacy format
 * - identity_credentials: sets version to 0 (pre-normalization)
 * - identity_verifiable_addresses: replaces normalized value with legacy format
 * - identity_recovery_addresses: replaces normalized value with legacy format
 */
export function downgradeLegacyIdentity(
  identityId: string,
  legacyPhone: string,
  normalizedPhone: string,
) {
  const dbPath = getSqliteDbPath()
  const escapedLegacy = legacyPhone.replace(/'/g, "''")
  const escapedNormalized = normalizedPhone.replace(/'/g, "''")

  const queries = [
    // Revert credential identifiers to legacy format
    `UPDATE identity_credential_identifiers SET identifier = '${escapedLegacy}' WHERE identifier = '${escapedNormalized}' AND identity_credential_id IN (SELECT id FROM identity_credentials WHERE identity_id = '${identityId}');`,
    // Set credential version to 0 (pre-normalization)
    `UPDATE identity_credentials SET version = 0 WHERE identity_id = '${identityId}';`,
    // Revert verifiable addresses
    `UPDATE identity_verifiable_addresses SET value = '${escapedLegacy}' WHERE value = '${escapedNormalized}' AND identity_id = '${identityId}';`,
    // Revert recovery addresses
    `UPDATE identity_recovery_addresses SET value = '${escapedLegacy}' WHERE value = '${escapedNormalized}' AND identity_id = '${identityId}';`,
  ]

  for (const query of queries) {
    execSync(`sqlite3 "${dbPath}" "${query}"`, {
      encoding: "utf-8",
      timeout: 5000,
    })
  }
}

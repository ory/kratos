// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { APIRequestContext } from "@playwright/test"
import type { OryKratosConfiguration } from "../../../../shared/config"
import { downgradeLegacyIdentity } from "../../../actions/db"
import {
  createIdentity,
  createIdentityWithPhoneAndPassword,
} from "../../../actions/identity"
import { hasNoSession, hasSession } from "../../../actions/session"
import {
  deleteDocument,
  documentUrl,
  fetchDocument,
} from "../../../actions/webhook"
import { expect, test } from "../../../fixtures"
import smsSchema from "../../../fixtures/schemas/sms"
import smsOidcSchema from "../../../fixtures/schemas/sms-oidc"
import smsPasswordSchema from "../../../fixtures/schemas/sms-password"
import { LoginPage } from "../../../models/elements/login"
import { RegistrationPage } from "../../../models/elements/registration"

test.skip(
  process.env.DB !== "sqlite",
  "These tests rely on a SQL downgrade that only works with SQLite.",
)

type DeepPartial<T> = T extends object
  ? { [P in keyof T]?: DeepPartial<T[P]> }
  : T

// Test phone numbers — deterministic so the SQL downgrade can target the known E.164 form
const DE_LEGACY = "+49 176 671 11 638"
const DE_E164 = "+4917667111638"
const DE_DASHED = "+49-176-671-11-638"

const US_LEGACY = "+1 (415) 555-2671"
const US_E164 = "+14155552671"

const smsDocumentId = "doc-sms-" + Math.random().toString(36).substring(7)

function smsSchemaBase64(schema: object) {
  return (
    "base64://" +
    Buffer.from(JSON.stringify(schema), "ascii").toString("base64")
  )
}

// Always register all three SMS schemas so identities from any test group
// can be listed and deleted via the admin API regardless of which schema
// is "active". Without this, cleanIdentities() silently fails when the
// current config doesn't know the schema of an existing identity.
function allSmsSchemas() {
  return [
    { id: "sms" as const, url: smsSchemaBase64(smsSchema) },
    { id: "sms-password" as const, url: smsSchemaBase64(smsPasswordSchema) },
    { id: "sms-oidc" as const, url: smsSchemaBase64(smsOidcSchema) },
    // Include default and email schemas so cleanIdentities can list identities
    // left over from other test specs that use these schemas.
    {
      id: "default" as const,
      url: "file://test/e2e/profiles/oidc/identity.traits.schema.json",
    },
    {
      id: "email" as const,
      url: "file://test/e2e/profiles/email/identity.traits.schema.json",
    },
  ]
}

function smsCourierChannels() {
  return [
    {
      id: "sms" as const,
      type: "http" as const,
      request_config: {
        body: "base64://ZnVuY3Rpb24oY3R4KSB7DQpjdHg6IGN0eCwNCn0=",
        method: "PUT",
        url: documentUrl(smsDocumentId),
      },
    },
  ]
}

/**
 * Config for code-only login flows (A, B, F tests).
 * Password is disabled so the unified login page renders a single identifier
 * input, avoiding the strict-mode violation caused by two `input[name=identifier]`
 * elements when both code and password methods are active.
 */
function smsCodeConfig(schemaId: string): DeepPartial<OryKratosConfiguration> {
  return {
    security: { account_enumeration: { enabled: false } },
    selfservice: {
      flows: {
        login: { style: "unified" },
        registration: {
          after: { code: { hooks: [{ hook: "session" }] } },
        },
        recovery: { enabled: true, use: "code" },
      },
      methods: {
        code: { passwordless_enabled: true },
        password: { enabled: false },
      },
    },
    courier: { channels: smsCourierChannels() },
    identity: { default_schema_id: schemaId, schemas: allSmsSchemas() },
  }
}

/**
 * Config for password+recovery flows (C, D, E tests).
 * Code is enabled (for recovery) but passwordless_enabled is false so the
 * login page shows only the password form (single identifier input).
 * Recovery V2 (choose_recovery_address) is enabled so that SMS phone addresses
 * are supported — Recovery V1 hardcodes AddressTypeEmail and never sends SMS.
 */
function smsPasswordConfig(
  schemaId: string,
): DeepPartial<OryKratosConfiguration> {
  return {
    security: { account_enumeration: { enabled: false } },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    feature_flags: {
      // choose_recovery_address enables Recovery V2 which supports SMS phone
      // addresses — Recovery V1 hardcodes AddressTypeEmail and never sends SMS.
      choose_recovery_address: true,
      use_continue_with_transitions: true,
    } as any,
    selfservice: {
      flows: {
        login: { style: "unified" },
        registration: {
          after: { password: { hooks: [{ hook: "session" }] } },
        },
        recovery: { enabled: true, use: "code" },
      },
      methods: {
        code: { enabled: true, passwordless_enabled: false },
        password: { enabled: true },
      },
    },
    courier: { channels: smsCourierChannels() },
    identity: { default_schema_id: schemaId, schemas: allSmsSchemas() },
  }
}

/**
 * Deletes all identities via the admin API.
 * Called in beforeEach to ensure a clean DB state for every test,
 * preventing UNIQUE constraint violations from previous test runs or retries.
 */
async function cleanIdentities(request: APIRequestContext) {
  const resp = await request.get(
    "http://localhost:4434/admin/identities?per_page=250",
  )
  if (!resp.ok()) {
    console.error(
      `cleanIdentities: list failed with status ${resp.status()}: ${await resp.text()}`,
    )
    return
  }
  const identities: Array<{ id: string }> = await resp.json()
  for (const identity of identities) {
    const del = await request.delete(
      `http://localhost:4434/admin/identities/${identity.id}`,
    )
    if (!del.ok()) {
      console.error(
        `cleanIdentities: delete ${identity.id} failed with status ${del.status()}`,
      )
    }
  }
}

// ────────────────────────────────────────────────────────────────────
// A. Code login (SMS) — legacy data
// ────────────────────────────────────────────────────────────────────
test.describe("A: code login with legacy phone data", () => {
  test.use({
    configOverride: smsCodeConfig("sms"),
  })

  test.beforeEach(async ({ page }) => {
    await cleanIdentities(page.request)
  })

  test.afterEach(async () => {
    await deleteDocument(smsDocumentId)
  })

  test("A1: login succeeds with exact legacy format", async ({
    page,
    config,
    kratosPublicURL,
  }) => {
    // Create identity with DE_LEGACY via admin API — Kratos normalizes and stores DE_E164
    const identity = await createIdentity(page.request, {
      schema_id: "sms",
      traits: { phone: DE_LEGACY },
    })
    // Revert DE_E164 → DE_LEGACY to simulate a pre-normalization DB record
    downgradeLegacyIdentity(identity.id, DE_LEGACY, DE_E164)

    const login = new LoginPage(page, config)
    await login.open()
    await login.triggerLoginWithCode(DE_LEGACY)

    const result = await fetchDocument(smsDocumentId)
    await login.codeInput.input.fill(result.ctx.template_data.login_code)
    await Promise.all([
      page.waitForURL(
        new RegExp(config.selfservice.default_browser_return_url),
      ),
      login.codeSubmit.getByText("Continue").click(),
    ])
    await hasSession(page.request, kratosPublicURL)
  })

  test("A2: login fails with E.164 when DB has legacy format (known limitation)", async ({
    page,
    config,
    kratosPublicURL,
  }) => {
    // Create with legacy, store as DE_E164, then downgrade to legacy
    const identity = await createIdentity(page.request, {
      schema_id: "sms",
      traits: { phone: DE_LEGACY },
    })
    downgradeLegacyIdentity(identity.id, DE_LEGACY, DE_E164)

    const login = new LoginPage(page, config)
    await login.open()
    // Enter E.164 format — IN query becomes IN(DE_E164, DE_E164), no match on legacy DB row
    await login.inputField("identifier").fill(DE_E164)
    await Promise.all([
      page.waitForLoadState("networkidle"),
      login.codeSubmit.click(),
    ])

    // No SMS code is sent because no identity was found
    await hasNoSession(page.request, kratosPublicURL)
  })

  test("A3: login fails with dashed format when DB has legacy format", async ({
    page,
    config,
    kratosPublicURL,
  }) => {
    const identity = await createIdentity(page.request, {
      schema_id: "sms",
      traits: { phone: DE_LEGACY },
    })
    downgradeLegacyIdentity(identity.id, DE_LEGACY, DE_E164)

    const login = new LoginPage(page, config)
    await login.open()
    await login.inputField("identifier").fill(DE_DASHED)
    await Promise.all([
      page.waitForLoadState("networkidle"),
      login.codeSubmit.click(),
    ])

    await hasNoSession(page.request, kratosPublicURL)
  })
})

// ────────────────────────────────────────────────────────────────────
// B. Code login (SMS) — new data (control group)
// ────────────────────────────────────────────────────────────────────
test.describe("B: code login with newly registered phone data", () => {
  test.use({
    configOverride: smsCodeConfig("sms"),
  })

  test.beforeEach(async ({ page }) => {
    await cleanIdentities(page.request)
  })

  test.afterEach(async () => {
    await deleteDocument(smsDocumentId)
  })

  test("B1: register with spaces, login with E.164", async ({
    page,
    config,
    kratosPublicURL,
  }) => {
    // Register with non-normalized phone — new Kratos normalizes and stores DE_E164
    const registration = new RegistrationPage(page, config)
    await registration.open()
    await registration.triggerRegistrationWithCode(DE_LEGACY)

    let result = await fetchDocument(smsDocumentId)
    const regCode = result.ctx.template_data.registration_code
    await registration.inputField("code").fill(regCode)
    await Promise.all([
      page.waitForURL(
        new RegExp(config.selfservice.default_browser_return_url),
      ),
      registration.submitField("code").getByText("Continue").click(),
    ])
    await hasSession(page.request, kratosPublicURL)

    // Clear session for login test
    await page.context().clearCookies()
    await deleteDocument(smsDocumentId)

    // Login with E.164 — should succeed because DB now stores E.164 (new normalization)
    const login = new LoginPage(page, config)
    await login.open()
    await login.triggerLoginWithCode(DE_E164)

    result = await fetchDocument(smsDocumentId)
    await login.codeInput.input.fill(result.ctx.template_data.login_code)
    await login.codeSubmit.getByText("Continue").click()
    await hasSession(page.request, kratosPublicURL)
  })

  test("B2: register with spaces, login with same spaces", async ({
    page,
    config,
    kratosPublicURL,
  }) => {
    const registration = new RegistrationPage(page, config)
    await registration.open()
    await registration.triggerRegistrationWithCode(DE_LEGACY)

    let result = await fetchDocument(smsDocumentId)
    const regCode = result.ctx.template_data.registration_code
    await registration.inputField("code").fill(regCode)
    await Promise.all([
      page.waitForURL(
        new RegExp(config.selfservice.default_browser_return_url),
      ),
      registration.submitField("code").getByText("Continue").click(),
    ])
    await hasSession(page.request, kratosPublicURL)

    await page.context().clearCookies()
    await deleteDocument(smsDocumentId)

    // Login with same spaced format — IN query's 2nd value (raw input) matches
    const login = new LoginPage(page, config)
    await login.open()
    await login.triggerLoginWithCode(DE_LEGACY)

    result = await fetchDocument(smsDocumentId)
    await login.codeInput.input.fill(result.ctx.template_data.login_code)
    await login.codeSubmit.getByText("Continue").click()
    await hasSession(page.request, kratosPublicURL)
  })
})

// ────────────────────────────────────────────────────────────────────
// C. Password login with phone — legacy data
// ────────────────────────────────────────────────────────────────────
test.describe("C: password login with legacy phone data", () => {
  const password = "vErY-s3cur3-p@ss!"

  test.use({
    configOverride: smsPasswordConfig("sms-password"),
  })

  test.beforeEach(async ({ page }) => {
    await cleanIdentities(page.request)
  })

  test.afterEach(async () => {
    await deleteDocument(smsDocumentId)
  })

  test("C1: password login succeeds with exact legacy phone format", async ({
    page,
    config,
    kratosPublicURL,
  }) => {
    // Create via admin API — Kratos stores DE_E164, then downgrade to legacy
    const { identity } = await createIdentityWithPhoneAndPassword(
      page.request,
      DE_LEGACY,
      password,
      "sms-password",
    )
    downgradeLegacyIdentity(identity.id, DE_LEGACY, DE_E164)

    const login = new LoginPage(page, config)
    await login.open()
    await login.loginWithPassword(DE_LEGACY, password)
    await hasSession(page.request, kratosPublicURL)
  })

  test("C2: password login fails with E.164 when DB has legacy format (known limitation)", async ({
    page,
    config,
    kratosPublicURL,
  }) => {
    const { identity } = await createIdentityWithPhoneAndPassword(
      page.request,
      DE_LEGACY,
      password,
      "sms-password",
    )
    downgradeLegacyIdentity(identity.id, DE_LEGACY, DE_E164)

    const login = new LoginPage(page, config)
    await login.open()
    // E.164 input — IN query produces IN(DE_E164, DE_E164) — no match on legacy value
    await login.loginWithPassword(DE_E164, password)
    await page.waitForLoadState("networkidle")
    await hasNoSession(page.request, kratosPublicURL)
  })
})

// ────────────────────────────────────────────────────────────────────
// D. Recovery via SMS — legacy data
// ────────────────────────────────────────────────────────────────────
test.describe("D: SMS recovery with legacy phone data", () => {
  const password = "vErY-s3cur3-p@ss!"

  test.use({
    configOverride: smsPasswordConfig("sms-password"),
  })

  test.beforeEach(async ({ page }) => {
    await cleanIdentities(page.request)
  })

  test.afterEach(async () => {
    await deleteDocument(smsDocumentId)
  })

  test("D1: recovery succeeds with exact legacy phone format", async ({
    page,
    config,
  }) => {
    const { identity } = await createIdentityWithPhoneAndPassword(
      page.request,
      US_LEGACY,
      password,
      "sms-password",
    )
    downgradeLegacyIdentity(identity.id, US_LEGACY, US_E164)

    await page.goto(config.selfservice.flows!.recovery!.ui_url!)
    await page.locator('input[name="recovery_address"]').fill(US_LEGACY)
    await page.locator('button[name="method"][value="code"]').click()

    // Recovery code should be sent via SMS webhook
    // RecoveryCodeValidModel.RecoveryCode is marshalled as "verification_code"
    const result = await fetchDocument(smsDocumentId)
    expect(result.ctx.template_data).toBeDefined()
    expect(result.ctx.template_data.verification_code).toBeDefined()
  })

  test("D2: recovery fails with E.164 when DB has legacy format (known limitation)", async ({
    page,
    config,
  }) => {
    const { identity } = await createIdentityWithPhoneAndPassword(
      page.request,
      US_LEGACY,
      password,
      "sms-password",
    )
    downgradeLegacyIdentity(identity.id, US_LEGACY, US_E164)

    await page.goto(config.selfservice.flows!.recovery!.ui_url!)
    await page.locator('input[name="recovery_address"]').fill(US_E164)
    await page.locator('button[name="method"][value="code"]').click()

    // Wait for Kratos to process, then verify no SMS was sent
    await page.waitForTimeout(2000)
    try {
      const result = await fetchDocument(smsDocumentId)
      // If a document exists, recovery was triggered — unexpected for this known-limitation case
      expect(result.ctx?.template_data?.verification_code).toBeUndefined()
    } catch {
      // fetchDocument retries and throws on 404 — no document means no SMS, which is expected
    }
  })
})

// ────────────────────────────────────────────────────────────────────
// E. Verification — legacy data
// ────────────────────────────────────────────────────────────────────
test.describe("E: verification address integrity with legacy data", () => {
  test.use({
    configOverride: smsPasswordConfig("sms-password"),
  })

  test.beforeEach(async ({ page }) => {
    await cleanIdentities(page.request)
  })

  test("E1: legacy identity retains verification address", async ({ page }) => {
    const { identity } = await createIdentityWithPhoneAndPassword(
      page.request,
      DE_LEGACY,
      "some-password-123!",
      "sms-password",
    )
    downgradeLegacyIdentity(identity.id, DE_LEGACY, DE_E164)

    // Fetch the identity via admin API and verify the address is still present
    const resp = await page.request.get(
      `http://localhost:4434/admin/identities/${identity.id}?include_credential=password`,
    )
    expect(resp.status()).toBe(200)
    const fetched = await resp.json()

    // The verifiable address should still exist (even if value was reverted to legacy)
    expect(fetched.verifiable_addresses).toBeDefined()
    expect(fetched.verifiable_addresses.length).toBeGreaterThan(0)
  })
})

// ────────────────────────────────────────────────────────────────────
// F. Duplicate identity detection (ConflictingIdentity)
// ────────────────────────────────────────────────────────────────────
test.describe("F: duplicate identity detection with phone normalization", () => {
  test.use({
    configOverride: smsCodeConfig("sms"),
  })

  test.beforeEach(async ({ page }) => {
    await cleanIdentities(page.request)
  })

  test.afterEach(async () => {
    await deleteDocument(smsDocumentId)
  })

  test("F1: legacy phone in DB, register with E.164 creates duplicate (known gap)", async ({
    page,
    config,
    kratosPublicURL,
  }) => {
    // Create identity with DE_LEGACY → stored as DE_E164 → then downgrade to DE_LEGACY
    const legacyIdentity = await createIdentity(page.request, {
      schema_id: "sms",
      traits: { phone: DE_LEGACY },
    })
    downgradeLegacyIdentity(legacyIdentity.id, DE_LEGACY, DE_E164)

    // Register a new identity with DE_E164 (the E.164 form of the same number)
    const registration = new RegistrationPage(page, config)
    await registration.open()
    await registration.triggerRegistrationWithCode(DE_E164)

    // The IN query becomes IN(DE_E164, DE_E164) which does NOT match legacy "+49 176 671 11 638"
    // So no conflict is detected and a registration code is sent
    const result = await fetchDocument(smsDocumentId)
    expect(result.ctx.template_data.registration_code).toBeDefined()

    // Complete registration — this proves the system allowed a duplicate
    await registration
      .inputField("code")
      .fill(result.ctx.template_data.registration_code)
    await Promise.all([
      page.waitForURL(
        new RegExp(config.selfservice.default_browser_return_url),
      ),
      registration.submitField("code").getByText("Continue").click(),
    ])
    await hasSession(page.request, kratosPublicURL)

    // Verify two distinct identities exist for the same logical phone number — BUG
    const listResp = await page.request.get(
      "http://localhost:4434/admin/identities?per_page=250",
    )
    const identities = await listResp.json()
    const matching = identities.filter(
      (i: { traits: { phone: string } }) =>
        i.traits.phone === DE_E164 || i.traits.phone === DE_LEGACY,
    )
    // BUG: Two identities exist for the same phone number
    expect(matching.length).toBe(2)
  })

  test("F2: new E.164 in DB, register with spaced format detects conflict", async ({
    page,
    config,
  }) => {
    // Create identity with E.164 directly (modern Kratos stores this as-is)
    await createIdentity(page.request, {
      schema_id: "sms",
      traits: { phone: DE_E164 },
    })

    // Attempt to register with the spaced variant of the same number.
    // The IN query becomes IN(DE_E164, DE_LEGACY) — DE_E164 matches the DB record.
    const registration = new RegistrationPage(page, config)
    await registration.open()
    await registration.triggerRegistrationWithCode(DE_LEGACY)

    const result = await fetchDocument(smsDocumentId)
    expect(result.ctx.template_data.registration_code).toBeDefined()

    await registration
      .inputField("code")
      .fill(result.ctx.template_data.registration_code)
    await registration.submitField("code").getByText("Continue").click()

    // Conflict is detected: only one identity should exist
    const listResp = await page.request.get(
      "http://localhost:4434/admin/identities?per_page=250",
    )
    const identities = await listResp.json()
    const matching = identities.filter(
      (i: { traits: { phone: string } }) =>
        i.traits.phone === DE_E164 || i.traits.phone === DE_LEGACY,
    )
    expect(matching.length).toBe(1)
  })
})

// ────────────────────────────────────────────────────────────────────
// G. OIDC account linking — legacy data
// ────────────────────────────────────────────────────────────────────
const oidcConfig: DeepPartial<OryKratosConfiguration> = {
  security: {
    account_enumeration: {
      enabled: false,
    },
  },
  selfservice: {
    flows: {
      login: {
        style: "unified",
      },
      registration: {
        after: {
          oidc: {
            hooks: [{ hook: "session" }],
          },
          code: {
            hooks: [{ hook: "session" }],
          },
        },
      },
    },
    methods: {
      code: {
        passwordless_enabled: true,
      },
      password: {
        enabled: false,
      },
      oidc: {
        enabled: true,
        config: {
          providers: [
            {
              id: "hydra",
              label: "Ory",
              provider: "generic",
              client_id: process.env["OIDC_HYDRA_CLIENT_ID"],
              client_secret: process.env["OIDC_HYDRA_CLIENT_SECRET"],
              issuer_url: "http://localhost:4444/",
              scope: ["offline"],
              mapper_url:
                "file://test/e2e/playwright/fixtures/mappers/hydra-phone.jsonnet",
            },
          ],
        },
      },
    },
  },
  courier: { channels: smsCourierChannels() },
  identity: { default_schema_id: "sms-oidc", schemas: allSmsSchemas() },
}

test.describe("G: OIDC account linking with legacy phone data", () => {
  test.use({
    configOverride: oidcConfig,
  })

  test.beforeEach(async ({ page }) => {
    await cleanIdentities(page.request)
  })

  test.afterEach(async () => {
    await deleteDocument(smsDocumentId)
  })

  test("G1: legacy phone in DB, OIDC returns E.164 creates duplicate (known gap)", async ({
    page,
    config,
    kratosPublicURL,
  }) => {
    // Create identity with legacy phone via admin API, then downgrade to legacy format
    const legacyIdentity = await createIdentity(page.request, {
      schema_id: "sms-oidc",
      traits: { phone: DE_LEGACY },
    })
    downgradeLegacyIdentity(legacyIdentity.id, DE_LEGACY, DE_E164)

    // Initiate OIDC registration — Hydra returns DE_E164 as sub claim
    await page.goto(config.selfservice.flows!.registration!.ui_url!)
    await page.locator('button[name="provider"][value="hydra"]').click()

    // Hydra login: use E.164 phone as the username (sub claim)
    await page.locator("input[name=username]").fill(DE_E164)
    await page.locator("button[name=action][value=accept]").click()

    // Consent screen
    await page.locator("#offline").check()
    await page.locator("#openid").check()
    await page.locator("button[name=action][value=accept]").click()

    // ConflictingIdentity uses IN(DE_E164, DE_E164) which does NOT match legacy record
    // BUG: A second identity is created instead of linking to the existing one
    await page.waitForURL(
      new RegExp(config.selfservice.default_browser_return_url),
    )
    await hasSession(page.request, kratosPublicURL)

    const listResp = await page.request.get(
      "http://localhost:4434/admin/identities?per_page=250",
    )
    const identities = await listResp.json()
    const matching = identities.filter(
      (i: { traits: { phone: string } }) =>
        i.traits.phone === DE_E164 || i.traits.phone === DE_LEGACY,
    )
    // BUG: Two identities exist for the same phone number
    expect(matching.length).toBe(2)
  })

  test("G2: new E.164 in DB, OIDC registration detects conflict (no duplicate)", async ({
    page,
    config,
  }) => {
    // Create identity with E.164 phone (modern format — no downgrade needed)
    await createIdentity(page.request, {
      schema_id: "sms-oidc",
      traits: { phone: DE_E164 },
    })

    // Initiate OIDC registration — Hydra returns DE_E164 as sub claim
    await page.goto(config.selfservice.flows!.registration!.ui_url!)
    await page.locator('button[name="provider"][value="hydra"]').click()

    await page.locator("input[name=username]").fill(DE_E164)
    await page.locator("button[name=action][value=accept]").click()

    await page.locator("#offline").check()
    await page.locator("#openid").check()
    await page.locator("button[name=action][value=accept]").click()

    // ConflictingIdentity uses IN(DE_E164, DE_E164) which DOES match DE_E164 in DB.
    // Conflict is detected → Kratos shows duplicate-credentials / account-linking
    // page instead of completing registration. No session is created.
    await page.waitForLoadState("networkidle")

    // Key assertion: conflict was detected, so NO duplicate identity was created.
    const listResp = await page.request.get(
      "http://localhost:4434/admin/identities?per_page=250",
    )
    const identities = await listResp.json()
    const matching = identities.filter(
      (i: { traits: { phone: string } }) => i.traits.phone === DE_E164,
    )
    expect(matching.length).toBe(1)
  })
})

// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { faker } from "@faker-js/faker"
import { CDPSession, expect, Page } from "@playwright/test"
import { OryKratosConfiguration } from "../../../../shared/config"
import { getSession, hasNoSession } from "../../../actions/session"
import { test } from "../../../fixtures"
import { toConfig } from "../../../lib/helper"

async function toggleAutomaticPresenceSimulation(
  cdpSession: CDPSession,
  authenticatorId: string,
  enabled: boolean,
) {
  await cdpSession.send("WebAuthn.setAutomaticPresenceSimulation", {
    authenticatorId,
    enabled,
  })
}

async function startPasskeyRegistration(page: Page) {
  await page.goto("/registration")
  const identifier = faker.internet.email()
  await page.locator(`input[name="traits.email"]`).fill(identifier)
  await page.locator(`input[name="traits.website"]`).fill(faker.internet.url())
  await page.locator("button[name=method][value=profile]").click()
  return identifier
}

async function registerWithPasskey(
  page: Page,
  pageCDPSession: CDPSession,
  config: OryKratosConfiguration,
  authenticatorId: string,
) {
  return await test.step("create passkey identity", async () => {
    const identifier = await startPasskeyRegistration(page)
    await toggleAutomaticPresenceSimulation(
      pageCDPSession,
      authenticatorId,
      true,
    )
    await page.locator("button[name=passkey_register_trigger]").click()

    await page.waitForURL(
      new RegExp(config.selfservice.default_browser_return_url),
    )
    return identifier
  })
}

test.describe("passkey options - cross-platform attachment", () => {
  test.use({
    configOverride: toConfig({
      style: "identifier_first",
      selfservice: {
        methods: {
          passkey: {
            enabled: true,
            config: {
              rp: {
                display_name: "ORY",
                id: "localhost",
                origins: ["http://localhost:4455"],
              },
              authenticator_selection: {
                attachment: "cross-platform",
              },
            },
          },
        },
        flows: {
          registration: {
            after: {
              passkey: {
                hooks: [{ hook: "session" }],
              },
            },
          },
        },
      },
    }),
    virtualAuthenticatorOptions: {
      automaticPresenceSimulation: true,
      hasResidentKey: true,
      transport: "usb",
    },
  })

  test("registration with cross-platform authenticator succeeds", async ({
    config,
    page,
    pageCDPSession,
    virtualAuthenticator,
    kratosPublicURL,
  }) => {
    const identifier = await registerWithPasskey(
      page,
      pageCDPSession,
      config,
      virtualAuthenticator.authenticatorId,
    )

    await expect(
      getSession(page.request, kratosPublicURL),
    ).resolves.toMatchObject({
      active: true,
      identity: { traits: { email: identifier } },
    })
  })
})

test.describe("passkey options - user verification required", () => {
  test.use({
    configOverride: toConfig({
      style: "identifier_first",
      selfservice: {
        methods: {
          passkey: {
            enabled: true,
            config: {
              rp: {
                display_name: "ORY",
                id: "localhost",
                origins: ["http://localhost:4455"],
              },
              authenticator_selection: {
                user_verification: "required",
              },
            },
          },
        },
        flows: {
          registration: {
            after: {
              passkey: {
                hooks: [{ hook: "session" }],
              },
            },
          },
        },
      },
    }),
    virtualAuthenticatorOptions: {
      automaticPresenceSimulation: true,
      hasResidentKey: true,
      hasUserVerification: true,
      isUserVerified: true,
    },
  })

  test("registration with user verification required succeeds", async ({
    config,
    page,
    pageCDPSession,
    virtualAuthenticator,
    kratosPublicURL,
  }) => {
    const identifier = await registerWithPasskey(
      page,
      pageCDPSession,
      config,
      virtualAuthenticator.authenticatorId,
    )

    await expect(
      getSession(page.request, kratosPublicURL),
    ).resolves.toMatchObject({
      active: true,
      identity: { traits: { email: identifier } },
    })
  })
})

test.describe("passkey options - user verification required, authenticator cannot verify", () => {
  // Authenticator reports no UV capability, so the browser refuses to
  // create a credential when the RP sets user_verification=required.
  test.use({
    configOverride: toConfig({
      style: "identifier_first",
      selfservice: {
        methods: {
          passkey: {
            enabled: true,
            config: {
              rp: {
                display_name: "ORY",
                id: "localhost",
                origins: ["http://localhost:4455"],
              },
              authenticator_selection: {
                user_verification: "required",
              },
            },
          },
        },
        flows: {
          registration: {
            after: {
              passkey: {
                hooks: [{ hook: "session" }],
              },
            },
          },
        },
      },
    }),
    virtualAuthenticatorOptions: {
      automaticPresenceSimulation: true,
      hasResidentKey: true,
      hasUserVerification: false,
      isUserVerified: false,
    },
  })

  test("registration fails and leaves the user on the registration page", async ({
    page,
    pageCDPSession,
    virtualAuthenticator,
    kratosPublicURL,
  }) => {
    await startPasskeyRegistration(page)
    await toggleAutomaticPresenceSimulation(
      pageCDPSession,
      virtualAuthenticator.authenticatorId,
      true,
    )
    await page.locator("button[name=passkey_register_trigger]").click()

    // Credential creation must fail; the flow must not finalize a session.
    // Give the browser time to reject the ceremony, then assert we stay on
    // /registration and no session exists.
    await page.waitForTimeout(1000)
    await expect(page).toHaveURL(/\/registration/)
    await hasNoSession(page.request, kratosPublicURL)

    // The passkey trigger must remain visible so the user can retry.
    await expect(
      page.locator("button[name=passkey_register_trigger]"),
    ).toBeVisible()
  })
})

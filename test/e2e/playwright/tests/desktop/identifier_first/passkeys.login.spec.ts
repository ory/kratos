// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { faker } from "@faker-js/faker"
import { CDPSession, expect, Page } from "@playwright/test"
import { OryKratosConfiguration } from "../../../../shared/config"
import { getSession } from "../../../actions/session"
import { test } from "../../../fixtures"
import { toConfig } from "../../../lib/helper"
import { LoginPage } from "../../../models/elements/login"

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

async function registerWithPasskey(
  page: Page,
  pageCDPSession: CDPSession,
  config: OryKratosConfiguration,
  authenticatorId: string,
  simulatePresence: boolean,
) {
  return await test.step("create webauthn identity", async () => {
    await page.goto("/registration")
    const identifier = faker.internet.email()
    await page.locator(`input[name="traits.email"]`).fill(identifier)
    await page
      .locator(`input[name="traits.website"]`)
      .fill(faker.internet.url())
    await page.locator("button[name=method][value=profile]").click()

    await toggleAutomaticPresenceSimulation(
      pageCDPSession,
      authenticatorId,
      true,
    )
    await page.locator("button[name=passkey_register_trigger]").click()

    await toggleAutomaticPresenceSimulation(
      pageCDPSession,
      authenticatorId,
      simulatePresence,
    )

    await page.waitForURL(
      new RegExp(config.selfservice.default_browser_return_url),
    )
    return identifier
  })
}

const passkeyConfig = {
  methods: {
    passkey: {
      enabled: true,
      config: {
        rp: {
          display_name: "ORY",
          id: "localhost",
          origins: ["http://localhost:4455"],
        },
      },
    },
  },
}

for (const mitigateEnumeration of [true, false]) {
  test.describe(`account enumeration protection ${
    mitigateEnumeration ? "on" : "off"
  }`, () => {
    test.use({
      configOverride: toConfig({
        mitigateEnumeration,
        style: "identifier_first",
        selfservice: passkeyConfig,
      }),
    })

    for (const simulatePresence of [true, false]) {
      test.describe(`${
        simulatePresence ? "with" : "without"
      } automatic presence proof`, () => {
        test.use({
          virtualAuthenticatorOptions: {
            automaticPresenceSimulation: simulatePresence,
            // hasResidentKey: simulatePresence,
          },
        })
        test("login", async ({
          config,
          page,
          kratosPublicURL,
          virtualAuthenticator,
          pageCDPSession,
        }) => {
          const identifier = await registerWithPasskey(
            page,
            pageCDPSession,
            config,
            virtualAuthenticator.authenticatorId,
            simulatePresence,
          )
          await page.context().clearCookies({})

          const login = new LoginPage(page, config)
          await login.open()

          if (!simulatePresence) {
            await login.submitIdentifierFirst(identifier)

            const passkeyLoginTrigger = page.locator(
              "button[name=passkey_login_trigger]",
            )
            await passkeyLoginTrigger.waitFor()

            await page.waitForLoadState("load")

            await toggleAutomaticPresenceSimulation(
              pageCDPSession,
              virtualAuthenticator.authenticatorId,
              true,
            )

            await passkeyLoginTrigger.click()

            await toggleAutomaticPresenceSimulation(
              pageCDPSession,
              virtualAuthenticator.authenticatorId,
              false,
            )
          }

          await page.waitForURL(
            new RegExp(config.selfservice.default_browser_return_url),
          )

          await expect(
            getSession(page.request, kratosPublicURL),
          ).resolves.toMatchObject({
            active: true,
            identity: {
              traits: {
                email: identifier,
              },
            },
          })
        })
      })
    }
  })
}

test.describe("without automatic presence simulation", () => {
  test.use({
    virtualAuthenticatorOptions: {
      automaticPresenceSimulation: false,
    },
    configOverride: toConfig({
      selfservice: passkeyConfig,
    }),
  })
  test("login with refresh", async ({
    page,
    config,
    kratosPublicURL,
    pageCDPSession,
    virtualAuthenticator,
  }) => {
    const identifier = await registerWithPasskey(
      page,
      pageCDPSession,
      config,
      virtualAuthenticator.authenticatorId,
      true,
    )

    const login = new LoginPage(page, config)
    // Due to resetting automatic presence simulating to "true" in the previous step,
    // opening the login page automatically triggers the passkey login
    await login.open()

    await page.waitForURL(
      new RegExp(config.selfservice.default_browser_return_url),
    )

    await expect(
      getSession(page.request, kratosPublicURL),
    ).resolves.toMatchObject({
      active: true,
      identity: {
        traits: {
          email: identifier,
        },
      },
    })

    await login.open({
      refresh: true,
    })

    await expect(
      page.locator('[data-testid="ui/message/1010003"]'),
      "show the refresh message",
    ).toBeVisible()

    const initialSession = await getSession(page.request, kratosPublicURL)

    const passkeyLoginTrigger = page.locator(
      "button[name=passkey_login_trigger]",
    )
    await passkeyLoginTrigger.waitFor()

    await page.waitForLoadState("load")

    await toggleAutomaticPresenceSimulation(
      pageCDPSession,
      virtualAuthenticator.authenticatorId,
      true,
    )

    await passkeyLoginTrigger.click()
    await page.waitForURL(
      new RegExp(config.selfservice.default_browser_return_url),
    )
    const newSession = await getSession(page.request, kratosPublicURL)

    expect(newSession.authentication_methods).toHaveLength(
      initialSession.authentication_methods.length + 1,
    )
  })
})

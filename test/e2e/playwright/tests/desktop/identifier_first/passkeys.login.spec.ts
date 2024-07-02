// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { faker } from "@faker-js/faker"
import { expect } from "@playwright/test"
import { getSession } from "../../../actions/session"
import { test } from "../../../fixtures"

test.use({
  addVirtualAuthenticator: true,
})

for (const mitigateEnumeration of [true, false]) {
  test.describe(`account enumeration protection ${
    mitigateEnumeration ? "on" : "off"
  }`, () => {
    test.use({
      configOverride: {
        security: {
          account_enumeration: {
            mitigate: mitigateEnumeration,
          },
        },
        selfservice: {
          flows: {
            login: {
              style: "identifier_first",
            },
            registration: {
              enable_legacy_one_step: false,
            },
          },
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
        },
      },
    })

    test("login", async ({ config, page, kratosPublicURL }) => {
      const identifier =
        await test.step("create webauthn identity", async () => {
          await page.goto("/registration")
          const identifier = faker.internet.email()
          await page.locator(`input[name="traits.email"]`).fill(identifier)
          await page
            .locator(`input[name="traits.website"]`)
            .fill(faker.internet.url())
          await page.locator("button[name=method][value=profile]").click()

          await page.locator("button[name=passkey_register_trigger]").click()

          return identifier
        })

      await page.goto("/login")

      await page.waitForURL(
        new RegExp(config.selfservice.default_browser_return_url),
      )

      const session = await getSession(page.request, kratosPublicURL)
      expect(session).toBeDefined()
      expect(session.active).toBe(true)
      expect(session.identity.traits.email).toBe(identifier)
    })
  })
}

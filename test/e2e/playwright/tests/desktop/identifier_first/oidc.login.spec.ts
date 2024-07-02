// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { faker } from "@faker-js/faker"
import { expect } from "@playwright/test"
import { getSession } from "../../../actions/session"
import { test } from "../../../fixtures"
import { toConfig } from "../../../lib/helper"

for (const mitigateEnumeration of [true, false]) {
  test.describe(`account enumeration protection ${
    mitigateEnumeration ? "on" : "off"
  }`, () => {
    test.use({
      configOverride: toConfig({ mitigateEnumeration }),
    })

    test("login", async ({ page, config, kratosPublicURL }) => {
      await page.goto("/login")

      await page.locator(`button[name=provider][value=hydra]`).click()

      await page
        .locator("input[name=username]")
        .fill(faker.internet.email({ provider: "ory.sh" }))
      await page.locator("button[name=action][value=accept]").click()
      await page.locator("#offline").check()
      await page.locator("#openid").check()

      await page.locator("input[name=website]").fill(faker.internet.url())

      await page.locator("button[name=action][value=accept]").click()

      await page.waitForURL(
        new RegExp(config.selfservice.default_browser_return_url),
      )

      const session = await getSession(page.request, kratosPublicURL)
      expect(session).toBeDefined()
      expect(session.active).toBe(true)
    })
  })
}

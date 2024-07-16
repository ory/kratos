// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { faker } from "@faker-js/faker"
import { expect, Page } from "@playwright/test"
import { getSession, hasSession } from "../../../actions/session"
import { test } from "../../../fixtures"
import { toConfig } from "../../../lib/helper"
import { LoginPage } from "../../../models/elements/login"
import { OryKratosConfiguration } from "../../../../shared/config"

async function loginHydra(page: Page) {
  return test.step("login with hydra", async () => {
    await page
      .locator("input[name=username]")
      .fill(faker.internet.email({ provider: "ory.sh" }))
    await page.locator("button[name=action][value=accept]").click()
    await page.locator("#offline").check()
    await page.locator("#openid").check()

    await page.locator("input[name=website]").fill(faker.internet.url())

    await page.locator("button[name=action][value=accept]").click()
  })
}

async function registerWithHydra(
  page: Page,
  config: OryKratosConfiguration,
  kratosPublicURL: string,
) {
  return await test.step("register", async () => {
    await page.goto("/registration")

    await page.locator(`button[name=provider][value=hydra]`).click()

    const email = faker.internet.email({ provider: "ory.sh" })
    await page.locator("input[name=username]").fill(email)
    await page.locator("#remember").check()
    await page.locator("button[name=action][value=accept]").click()
    await page.locator("#offline").check()
    await page.locator("#openid").check()

    await page.locator("input[name=website]").fill(faker.internet.url())

    await page.locator("button[name=action][value=accept]").click()
    await page.waitForURL(
      new RegExp(config.selfservice.default_browser_return_url),
    )
    await page.context().clearCookies({
      domain: new URL(kratosPublicURL).hostname,
    })

    await expect(
      getSession(page.request, kratosPublicURL),
    ).rejects.toThrowError()
    return email
  })
}

for (const mitigateEnumeration of [true, false]) {
  test.describe(`account enumeration protection ${
    mitigateEnumeration ? "on" : "off"
  }`, () => {
    test.use({
      configOverride: toConfig({
        mitigateEnumeration,
        selfservice: {
          methods: {
            password: {
              enabled: true,
            },
          },
        },
      }),
    })

    test("login", async ({ page, config, kratosPublicURL }) => {
      const login = new LoginPage(page, config)
      await login.open()

      await page.locator(`button[name=provider][value=hydra]`).click()

      await loginHydra(page)

      await page.waitForURL(
        new RegExp(config.selfservice.default_browser_return_url),
      )

      await hasSession(page.request, kratosPublicURL)
    })

    test("oidc sign in on second step", async ({
      page,
      config,
      kratosPublicURL,
    }) => {
      const email = await registerWithHydra(page, config, kratosPublicURL)

      const login = new LoginPage(page, config)
      await login.open()

      await login.submitIdentifierFirst(email)

      // If account enumeration is mitigated, we should see the password method,
      // because the identity has not set up a password
      await expect(
        page.locator('button[name="method"][value="password"]'),
        "hide the password method",
      ).toBeVisible({ visible: mitigateEnumeration })

      await page.locator(`button[name=provider][value=hydra]`).click()

      await loginHydra(page)

      await page.waitForURL(
        new RegExp(config.selfservice.default_browser_return_url),
      )

      const session = await getSession(page.request, kratosPublicURL)
      expect(session).toBeDefined()
      expect(session.active).toBe(true)
    })
  })
}

test("login with refresh", async ({ page, config, kratosPublicURL }) => {
  await registerWithHydra(page, config, kratosPublicURL)

  const login = new LoginPage(page, config)

  const initialSession = await test.step("initial login", async () => {
    await login.open()
    await page.locator(`button[name=provider][value=hydra]`).click()

    await loginHydra(page)

    await page.waitForURL(
      new RegExp(config.selfservice.default_browser_return_url),
    )
    return await getSession(page.request, kratosPublicURL)
  })

  // This is required, because OIDC issues a new session on refresh (TODO), and MySQL does not store sub second timestamps, so we need to wait a bit
  await page.waitForTimeout(1000)
  await test.step("refresh login", async () => {
    await login.open({
      refresh: true,
    })

    await expect(
      page.locator('[data-testid="ui/message/1010003"]'),
      "show the refresh message",
    ).toBeVisible()

    await page.locator(`button[name=provider][value=hydra]`).click()

    await loginHydra(page)

    await page.waitForURL(
      new RegExp(config.selfservice.default_browser_return_url),
    )
    const newSession = await getSession(page.request, kratosPublicURL)
    // expect(newSession.authentication_methods).toHaveLength(
    //   initialSession.authentication_methods.length + 1,
    // )
    expect(newSession.authenticated_at).not.toBe(
      initialSession.authenticated_at,
    )
  })
})

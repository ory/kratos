// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { faker } from "@faker-js/faker"
import { Session } from "@ory/kratos-client"
import { APIRequestContext, expect } from "@playwright/test"
import { test } from "../../../fixtures"

async function toSession(
  r: APIRequestContext,
  kratosPublicURL: string,
): Promise<Session> {
  const resp = await r.get(kratosPublicURL + "/sessions/whoami", {
    failOnStatusCode: true,
  })
  return resp.json()
}

test.use({
  configOverride: {
    selfservice: {
      default_browser_return_url: "http://localhost:4455/welcome",
      flows: {
        login: {
          style: "identifier_first",
        },
      },
    },
  },
})
test.describe("password", () => {
  test("login with password", async ({
    page,
    // projectFrontendClient,
    identity,
    config,
    kratosPublicURL,
  }) => {
    await page.goto("/login")

    const identifier = page.locator("input[name=identifier]")
    await expect(identifier).toBeVisible()
    await identifier.fill(identity.email)

    await page.locator("button[name=method][value=identifier_first]").click()

    const passwordInput = page.locator("input[name=password]")
    await expect(passwordInput).toBeVisible()
    await passwordInput.fill(identity.password)
    await page.locator("button[name=method][value=password]").click()

    await page.waitForURL(
      new RegExp(config.selfservice.default_browser_return_url),
    )

    const session = await toSession(page.request, kratosPublicURL)
    expect(session).toBeDefined()
    expect(session.active).toBe(true)

    await test.step("refresh", async () => {
      await page.goto("/login?refresh=true")

      const passwordInput = page.locator("input[name=password]")
      await expect(passwordInput).toBeVisible()
      await passwordInput.fill(identity.password)
      await page.locator("button[name=method][value=password]").click()
      await page.waitForURL(
        new RegExp(config.selfservice.default_browser_return_url),
      )
      const refreshedSession = await toSession(page.request, kratosPublicURL)
      expect(refreshedSession).toBeDefined()
      expect(refreshedSession.active).toBe(true)

      expect(session.expires_at).not.toBe(refreshedSession.expires_at)
    })
  })

  test("login with wrong password fails", async ({
    page,
    identity,
    kratosPublicURL,
  }) => {
    await page.goto("/login")

    const identifier = page.locator("input[name=identifier]")
    await expect(identifier).toBeVisible()
    await identifier.fill(identity.email)

    await page.locator("button[name=method][value=identifier_first]").click()

    const passwordInput = page.locator("input[name=password]")
    await expect(passwordInput).toBeVisible()
    await passwordInput.fill("aoksndon")
    await page.locator("button[name=method][value=password]").click()

    // await expect(page.getByTestId("ory-message-4000006")).toBeVisible()
    await expect(page.getByTestId("ui/message/4000006")).toBeVisible()

    await expect(toSession(page.request, kratosPublicURL)).rejects.toThrow()
  })
})

test.describe("oidc", () => {
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

    const session = await toSession(page.request, kratosPublicURL)
    expect(session).toBeDefined()
    expect(session.active).toBe(true)
  })
})

test.describe("passkeys", () => {
  test.use({
    addVirtualAuthenticator: true,
  })
  test.use({
    configOverride: async ({ page }, use) => {
      await use({
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
      })
    },
  })
  test("login", async ({ config, page, kratosPublicURL }) => {
    const identifier = await test.step("create webauthn identity", async () => {
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

    const session = await toSession(page.request, kratosPublicURL)
    expect(session).toBeDefined()
    expect(session.active).toBe(true)
  })
})

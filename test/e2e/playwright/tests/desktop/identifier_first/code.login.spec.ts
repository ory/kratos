// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { expect } from "@playwright/test"
import { search } from "../../../actions/mail"
import { getSession, hasNoSession, hasSession } from "../../../actions/session"
import { test } from "../../../fixtures"
import { extractCode, toConfig } from "../../../lib/helper"
import { LoginPage } from "../../../models/elements/login"

test.describe("account enumeration protection off", () => {
  test.use({
    configOverride: toConfig({
      style: "identifier_first",
      mitigateEnumeration: false,
      selfservice: {
        methods: {
          code: {
            passwordless_enabled: true,
          },
        },
      },
    }),
  })

  test("login fails because user does not exist", async ({ page, config }) => {
    const login = new LoginPage(page, config)
    await login.open()

    await login.submitIdentifierFirst("i@donot.exist")

    await expect(
      page.locator('[data-testid="ui/message/4000037"]'),
      "expect account not exist message to be shown",
    ).toBeVisible()
  })

  test("login with wrong code fails", async ({
    page,
    identity,
    kratosPublicURL,
    config,
  }) => {
    const login = new LoginPage(page, config)
    await login.open()

    await login.triggerLoginWithCode(identity.email)

    await login.codeInput.input.fill("123123")

    await login.codeSubmit.getByText("Continue").click()

    await hasNoSession(page.request, kratosPublicURL)
    await expect(
      page.locator('[data-testid="ui/message/4010008"]'),
      "expect to be shown a wrong code error",
    ).toBeVisible()
  })

  test("login succeeds", async ({
    page,
    identity,
    config,
    kratosPublicURL,
  }) => {
    const login = new LoginPage(page, config)
    await login.open()

    await login.triggerLoginWithCode(identity.email)

    const mails = await search({ query: identity.email, kind: "to" })
    expect(mails).toHaveLength(1)

    const code = extractCode(mails[0])

    await login.codeInput.input.fill(code)

    await login.codeSubmit.getByText("Continue").click()

    await hasSession(page.request, kratosPublicURL)
  })
})

test.describe("account enumeration protection on", () => {
  test.use({
    configOverride: toConfig({
      style: "identifier_first",
      mitigateEnumeration: true,
      selfservice: {
        methods: {
          password: {
            enabled: false,
          },
          code: {
            passwordless_enabled: true,
          },
        },
      },
    }),
  })

  test("login fails because user does not exist", async ({ page, config }) => {
    const login = new LoginPage(page, config)
    await login.open()

    await login.submitIdentifierFirst("i@donot.exist")

    await expect(
      page.locator('button[name="method"][value="code"]'),
      "expect to show the code form",
    ).toBeVisible()
  })

  test("login with wrong code fails", async ({
    page,
    identity,
    kratosPublicURL,
    config,
  }) => {
    const login = new LoginPage(page, config)
    await login.open()

    await login.triggerLoginWithCode(identity.email)

    await login.codeInput.input.fill("123123")

    await login.codeSubmit.getByText("Continue").click()

    await hasNoSession(page.request, kratosPublicURL)
    await expect(
      page.locator('[data-testid="ui/message/4010008"]'),
      "expect to be shown a wrong code error",
    ).toBeVisible()
  })

  test("login succeeds", async ({
    page,
    identity,
    config,
    kratosPublicURL,
  }) => {
    const login = new LoginPage(page, config)
    await login.open()

    await login.triggerLoginWithCode(identity.email)

    const mails = await search({ query: identity.email, kind: "to" })
    expect(mails).toHaveLength(1)

    const code = extractCode(mails[0])

    await login.codeInput.input.fill(code)

    await login.codeSubmit.getByText("Continue").click()

    await hasSession(page.request, kratosPublicURL)
  })
})

test.describe(() => {
  test.use({
    configOverride: toConfig({
      style: "identifier_first",
      mitigateEnumeration: false,
      selfservice: {
        methods: {
          code: {
            passwordless_enabled: true,
          },
        },
      },
    }),
  })
  test("refresh", async ({ page, identity, config, kratosPublicURL }) => {
    const login = new LoginPage(page, config)

    const [initialSession, initialCode] =
      await test.step("initial login", async () => {
        await login.open()
        await login.triggerLoginWithCode(identity.email)

        const mails = await search({ query: identity.email, kind: "to" })
        expect(mails).toHaveLength(1)

        const code = extractCode(mails[0])

        await login.codeInput.input.fill(code)

        await login.codeSubmit.getByText("Continue").click()

        const session = await getSession(page.request, kratosPublicURL)
        expect(session).toBeDefined()
        expect(session.active).toBe(true)
        return [session, code]
      })

    await login.open({
      refresh: true,
    })
    await login.inputField("identifier").fill(identity.email)
    await login.submit("code")

    const mails = await search({
      query: identity.email,
      kind: "to",
      filter: (m) => !m.html.includes(initialCode),
    })
    expect(mails).toHaveLength(1)

    const code = extractCode(mails[0])

    await login.codeInput.input.fill(code)

    await login.codeSubmit.getByText("Continue").click()
    await page.waitForURL(
      new RegExp(config.selfservice.default_browser_return_url),
    )

    const newSession = await getSession(page.request, kratosPublicURL)
    expect(newSession).toBeDefined()
    expect(newSession.active).toBe(true)

    const initDate = Date.parse(initialSession.authenticated_at)
    const newDate = Date.parse(newSession.authenticated_at)
    expect(newDate).toBeGreaterThanOrEqual(initDate)
  })
})

// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { expect } from "@playwright/test"
import { search } from "../../../actions/mail"
import { hasNoSession, hasSession } from "../../../actions/session"
import { test } from "../../../fixtures"
import { extractCode, toConfig } from "../../../lib/helper"
import { LoginPage } from "../../../models/elements/login"

test.describe.parallel("account enumeration protection off", () => {
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
    // projectFrontendClient,
    identity,
    config,
    kratosPublicURL,
  }) => {
    const login = new LoginPage(page, config)
    await login.open()

    await login.triggerLoginWithCode(identity.email)

    const mails = await search(identity.email, "to")
    expect(mails).toHaveLength(1)

    const code = extractCode(mails[0])

    await login.codeInput.input.fill(code)

    await login.codeSubmit.getByText("Continue").click()

    await hasSession(page.request, kratosPublicURL)
  })
  // TODO: add refresh tests
})

test.describe("account enumeration protection on", () => {
  test.use({
    configOverride: toConfig({
      style: "identifier_first",
      mitigateEnumeration: true,
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
    // projectFrontendClient,
    identity,
    config,
    kratosPublicURL,
  }) => {
    const login = new LoginPage(page, config)
    await login.open()

    await login.triggerLoginWithCode(identity.email)

    const mails = await search(identity.email, "to")
    expect(mails).toHaveLength(1)

    const code = extractCode(mails[0])

    await login.codeInput.input.fill(code)

    await login.codeSubmit.getByText("Continue").click()

    await hasSession(page.request, kratosPublicURL)
  })

  // TODO: add refresh tests
})

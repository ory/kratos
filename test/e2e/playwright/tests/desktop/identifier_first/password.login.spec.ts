// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { expect } from "@playwright/test"
import { loginWithPassword } from "../../../actions/login"
import { getSession, hasNoSession, hasSession } from "../../../actions/session"
import { test } from "../../../fixtures"
import { toConfig } from "../../../lib/helper"
import { LoginPage } from "../../../models/elements/login"

// These can run in parallel because they use the same config.
test.describe("account enumeration protection off", () => {
  test.use({
    configOverride: toConfig({
      style: "identifier_first",
      mitigateEnumeration: false,
      selfservice: {
        methods: {
          password: {
            enabled: true,
          },
          code: {
            passwordless_enabled: false,
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

  test("login with wrong password fails", async ({
    page,
    identity,
    kratosPublicURL,
    config,
  }) => {
    const login = new LoginPage(page, config)
    await login.open()

    await login.loginWithPassword(identity.email, "wrong-password")
    await login.isReady()

    await hasNoSession(page.request, kratosPublicURL)
    await expect(
      page.locator('[data-testid="ui/message/4000006"]'),
      "expect to be shown a credentials do not exist error",
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

    await login.inputField("identifier").fill(identity.email)
    await login.submit("identifier_first", {
      waitForURL: new RegExp(config.selfservice.flows.login.ui_url),
    })

    await login.inputField("password").fill(identity.password)
    await login.submit("password", {
      waitForURL: new RegExp(config.selfservice.default_browser_return_url),
    })

    await hasSession(page.request, kratosPublicURL)
  })

  test("login with refresh", async ({
    page,
    config,
    identity,
    kratosPublicURL,
  }) => {
    await loginWithPassword(
      {
        password: identity.password,
        traits: {
          email: identity.email,
        },
      },
      page.request,
      kratosPublicURL,
    )

    const login = new LoginPage(page, config)
    await login.open({
      refresh: true,
    })

    await expect(
      page.locator('[data-testid="ui/message/1010003"]'),
      "show the refresh message",
    ).toBeVisible()

    const initialSession = await getSession(page.request, kratosPublicURL)
    await login.inputField("password").fill(identity.password)
    await login.submit("password", {
      waitForURL: new RegExp(config.selfservice.default_browser_return_url),
    })

    const newSession = await getSession(page.request, kratosPublicURL)

    expect(newSession.authentication_methods).toHaveLength(
      initialSession.authentication_methods.length + 1,
    )
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
            enabled: true,
          },
          code: {
            passwordless_enabled: false,
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
      page.locator('button[name="method"][value="password"]'),
      "expect to show the password form",
    ).toBeVisible()
  })

  test("login with wrong password fails", async ({
    page,
    identity,
    kratosPublicURL,
    config,
  }) => {
    const login = new LoginPage(page, config)
    await login.open()

    await login.loginWithPassword(identity.email, "wrong-password")
    await login.isReady()

    await hasNoSession(page.request, kratosPublicURL)
    await expect(
      page.locator('[data-testid="ui/message/4000006"]'),
      "expect to be shown a credentials do not exist error",
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

    await login.inputField("identifier").fill(identity.email)
    await login.submit("identifier_first", {
      waitForURL: new RegExp(config.selfservice.flows.login.ui_url),
    })

    await login.inputField("password").fill(identity.password)
    await login.submit("password", {
      waitForURL: new RegExp(config.selfservice.default_browser_return_url),
    })

    await hasSession(page.request, kratosPublicURL)
  })

  test("login with refresh", async ({
    page,
    config,
    identity,
    kratosPublicURL,
  }) => {
    await loginWithPassword(
      {
        password: identity.password,
        traits: {
          email: identity.email,
        },
      },
      page.request,
      kratosPublicURL,
    )

    const login = new LoginPage(page, config)
    await login.open({
      refresh: true,
    })

    await expect(
      page.locator('[data-testid="ui/message/1010003"]'),
      "show the refresh message",
    ).toBeVisible()

    const initialSession = await getSession(page.request, kratosPublicURL)

    await login.inputField("password").fill(identity.password)
    await login.submit("password", {
      waitForURL: new RegExp(config.selfservice.default_browser_return_url),
    })

    const newSession = await getSession(page.request, kratosPublicURL)

    expect(newSession.authentication_methods).toHaveLength(
      initialSession.authentication_methods.length + 1,
    )
  })
})

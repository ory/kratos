// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { expect } from "@playwright/test"
import { search } from "../../../actions/mail"
import { getSession, hasNoSession, hasSession } from "../../../actions/session"
import { test } from "../../../fixtures"
import { extractCode, toConfig } from "../../../lib/helper"
import { LoginPage } from "../../../models/elements/login"
import { SettingsPage } from "../../../models/elements/settings"
import { logoutUrl } from "../../../actions/login"

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

for (const tc of [
  {
    name: "do fast login when only code method enabled and configured, mitigation off",
    methodsEnabled: ["code"],
    methodsConfigured: "code",
    mitigation: false,
    expectFastLogin: true,
  },
  {
    name: "do fast login when only code method enabled and configured, mitigation on",
    methodsEnabled: ["code"],
    methodsConfigured: "code",
    mitigation: true,
    expectFastLogin: true,
  },
  {
    name: "do not fast login when multiple methods enabled, all configured, mitigation off",
    methodsEnabled: ["password", "code"],
    methodsConfigured: "all",
    mitigation: false,
    expectFastLogin: false,
  },
  {
    name: "do not fast login when multiple methods enabled, all configured, mitigation on",
    methodsEnabled: ["password", "code"],
    methodsConfigured: "all",
    mitigation: true,
    expectFastLogin: false,
  },
  {
    name: "do fast login when multiple methods enabled, only code configured, mitigation off",
    methodsEnabled: ["password", "code"],
    methodsConfigured: "code",
    mitigation: false,
    expectFastLogin: true,
  },
  {
    name: "do not fast login when multiple methods enabled, only code configured, mitigation on",
    methodsEnabled: ["password", "code"],
    methodsConfigured: "code",
    mitigation: true,
    expectFastLogin: false,
  },
]) {
  test.describe(`account enumeration protection ${
    tc.mitigation ? "on" : "off"
  }`, () => {
    test.use({
      configOverride: toConfig({
        style: "identifier_first",
        mitigateEnumeration: tc.mitigation,
        selfservice: {
          methods: {
            oidc: {
              enabled: false,
            },
            password: {
              enabled: tc.methodsEnabled.includes("password"),
            },
            code: {
              passwordless_enabled: tc.methodsEnabled.includes("code"),
            },
          },
        },
      }),
    })

    test(
      tc.name,
      async ({ page, config, identity, identityWithoutPassword }) => {
        const id =
          tc.methodsConfigured === "all" ? identity : identityWithoutPassword

        const login = new LoginPage(page, config)
        await login.open()

        await login.submitIdentifierFirst(id.email)

        if (tc.expectFastLogin) {
          await expect(login.submitPassword).toBeHidden()
          await expect(
            page.locator('[data-testid="ui/message/1010014"]'),
            "expect code sent message to be shown",
          ).toBeVisible()
          await expect(login.codeSubmit).toBeVisible()
        } else {
          await expect(login.submitPassword).toBeVisible()
          await expect(login.codeSubmit).toBeVisible()
        }
      },
    )
  })
}

test.describe(() => {
  test.use({
    configOverride: toConfig({
      style: "identifier_first",
      mitigateEnumeration: false,
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

test.describe("second factor", () => {
  for (const tc of [
    {
      name: "do fast login when only code method enabled and configured",
      methodsEnabled: ["code"],
      methodsConfigured: "code",
      expectFastLogin: true,
    },
    {
      name: "do not fast login when multiple methods enabled, all configured",
      methodsEnabled: ["totp", "code"],
      methodsConfigured: "all",
      expectFastLogin: false,
    },
    {
      name: "do fast login when multiple methods enabled, only code configured",
      methodsEnabled: ["totp", "code"],
      methodsConfigured: "code",
      expectFastLogin: true,
    },
  ]) {
    test.describe(`code mfa ${
      tc.expectFastLogin ? "with" : "without"
    } fast login`, () => {
      test.use({
        configOverride: toConfig({
          style: "identifier_first",
          selfservice: {
            methods: {
              oidc: {
                enabled: false,
              },
              password: {
                enabled: true,
              },
              totp: {
                enabled: tc.methodsEnabled.includes("totp"),
              },
              code: {
                passwordless_enabled: false,
                mfa_enabled: tc.methodsEnabled.includes("code"),
              },
            },
          },
        }),
      })

      test(tc.name, async ({ page, config, identity, kratosPublicURL }) => {
        const login = new LoginPage(page, config)

        if (tc.methodsConfigured === "all") {
          await test.step("setup TOTP", async () => {
            await login.open()
            await login.loginWithPassword(identity.email, identity.password)

            const mails = await search({ query: identity.email, kind: "to" })
            expect(mails).toHaveLength(1)
            const code = extractCode(mails[0])
            await login.codeInput.input.fill(code)
            await login.codeSubmit.getByText("Continue").click()

            const settings = new SettingsPage(page, config)
            await settings.open()

            await settings.setupTotp()

            await page.goto(await logoutUrl(page.request, kratosPublicURL))
          })
        }

        test.step("first factor authentication", async () => {
          await login.open()
          await login.loginWithPassword(identity.email, identity.password)
        })

        if (tc.expectFastLogin) {
          await expect(login.totpInput.input).toBeHidden()
          await expect(
            page.locator('[data-testid="ui/message/1010014"]'),
            "expect code sent message to be shown",
          ).toBeVisible()
          await expect(login.codeSubmit).toBeVisible()
        } else {
          await expect(login.totpInput.input).toBeVisible()
          await expect(login.codeSubmitMfa).toBeVisible()
        }
      })
    })
  }
})

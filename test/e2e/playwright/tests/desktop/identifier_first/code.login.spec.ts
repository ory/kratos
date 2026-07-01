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
      featureFlags: {
        refresh_login_choose_address: true,
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
    // On a refresh login the identity is fixed by the session, so the screen
    // shows a "Send code to <address>" button instead of an identifier field
    // (ory/kratos#4194).
    await expect(login.inputField("identifier")).toBeHidden()
    await login.codeSubmitMfa.click()

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
            page.locator('[data-testid="ui/message/1010025"]'),
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

test.describe("second factor refresh", () => {
  // TOTP is enabled alongside code so the code method does not fast-login as the
  // only second factor. That keeps the flow on the regular handler path, which
  // renders the "Send code to <address>" button this test exercises.
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
            enabled: true,
          },
          code: {
            passwordless_enabled: false,
            mfa_enabled: true,
          },
        },
      },
    }),
  })

  test("refreshes an aal2 session through the code address button", async ({
    page,
    config,
    identity,
    kratosPublicURL,
  }) => {
    const login = new LoginPage(page, config)
    const seenCodes: string[] = []

    await test.step("set up TOTP as a second 2FA method", async () => {
      await login.open()
      await login.loginWithPassword(identity.email, identity.password)

      // The verified email already works as a code second factor, so with
      // `highest_available` the flow fast-logs in to aal2 with code. Complete
      // that before opening settings to set up TOTP.
      const mails = await search({
        query: identity.email,
        kind: "to",
        filter: (m) => !seenCodes.some((c) => m.html.includes(c)),
      })
      const code = extractCode(mails[0])
      seenCodes.push(code)
      await login.codeInput.input.fill(code)
      await login.codeSubmit.getByText("Continue").click()

      const settings = new SettingsPage(page, config)
      await settings.open()
      await settings.setupTotp()

      await page.goto(await logoutUrl(page.request, kratosPublicURL))
    })

    const initialSession =
      await test.step("log in to aal2 with the code second factor", async () => {
        await login.open()
        await login.loginWithPassword(identity.email, identity.password)

        // With TOTP and code both configured, the flow escalates to aal2 and
        // renders both factors instead of fast-logging in.
        await expect(login.totpInput.input).toBeVisible()
        await expect(login.codeSubmitMfa).toBeVisible()

        await login.codeSubmitMfa.click()

        const mails = await search({
          query: identity.email,
          kind: "to",
          filter: (m) => !seenCodes.some((c) => m.html.includes(c)),
        })
        const code = extractCode(mails[0])
        seenCodes.push(code)

        await login.codeInput.input.fill(code)
        await login.codeSubmit.getByText("Continue").click()

        const session = await getSession(page.request, kratosPublicURL)
        expect(session.active).toBe(true)
        expect(session.authenticator_assurance_level).toBe("aal2")
        return session
      })

    await test.step("refresh the aal2 session", async () => {
      await login.open({
        refresh: true,
        aal: "aal2",
      })

      // On a refresh + aal2 flow the second-factor hydrator wins over the
      // first-factor refresh branch: it renders "Send code to <address>"
      // buttons and never the hidden first-factor identifier field
      // (ory/kratos#4194).
      await expect(login.codeSubmitMfa).toBeVisible()
      await expect(login.inputField("identifier")).toBeHidden()

      await login.codeSubmitMfa.click()

      const mails = await search({
        query: identity.email,
        kind: "to",
        filter: (m) => !seenCodes.some((c) => m.html.includes(c)),
      })
      expect(mails).toHaveLength(1)
      const code = extractCode(mails[0])

      await login.codeInput.input.fill(code)
      await login.codeSubmit.getByText("Continue").click()
      await page.waitForURL(
        new RegExp(config.selfservice.default_browser_return_url),
      )

      const refreshedSession = await getSession(page.request, kratosPublicURL)
      expect(refreshedSession.active).toBe(true)
      expect(refreshedSession.authenticator_assurance_level).toBe("aal2")

      const initialAuthAt = Date.parse(initialSession.authenticated_at)
      const refreshedAuthAt = Date.parse(refreshedSession.authenticated_at)
      expect(refreshedAuthAt).toBeGreaterThanOrEqual(initialAuthAt)
    })
  })

  test("rejects a foreign address on an aal2 refresh", async ({
    page,
    config,
    identity,
    kratosPublicURL,
  }) => {
    const login = new LoginPage(page, config)
    const seenCodes: string[] = []

    await test.step("set up TOTP as a second 2FA method", async () => {
      await login.open()
      await login.loginWithPassword(identity.email, identity.password)

      // Complete the code second factor that fast-login auto-sends, then set
      // up TOTP so later flows render the address button instead of fast-login.
      const mails = await search({
        query: identity.email,
        kind: "to",
        filter: (m) => !seenCodes.some((c) => m.html.includes(c)),
      })
      const code = extractCode(mails[0])
      seenCodes.push(code)
      await login.codeInput.input.fill(code)
      await login.codeSubmit.getByText("Continue").click()

      const settings = new SettingsPage(page, config)
      await settings.open()
      await settings.setupTotp()

      await page.goto(await logoutUrl(page.request, kratosPublicURL))
    })

    const aal2Session =
      await test.step("log in to aal2 with the code second factor", async () => {
        await login.open()
        await login.loginWithPassword(identity.email, identity.password)

        await login.codeSubmitMfa.click()

        const mails = await search({
          query: identity.email,
          kind: "to",
          filter: (m) => !seenCodes.some((c) => m.html.includes(c)),
        })
        const code = extractCode(mails[0])
        seenCodes.push(code)
        await login.codeInput.input.fill(code)
        await login.codeSubmit.getByText("Continue").click()

        const session = await getSession(page.request, kratosPublicURL)
        expect(session.authenticator_assurance_level).toBe("aal2")
        return session
      })

    await test.step("tampering the address with a foreign identifier is rejected", async () => {
      await login.open({
        refresh: true,
        aal: "aal2",
      })

      // Tamper the server-rendered address button so it posts a foreign
      // identifier. A refresh login must not send a code to it
      // (ory/kratos#4194).
      const foreignEmail = "not-" + identity.email
      await login.codeSubmitMfa.evaluate(
        (el, value) => el.setAttribute("value", value),
        foreignEmail,
      )
      await login.codeSubmitMfa.click()

      await expect(
        page.locator('[data-testid="ui/message/4000035"]'),
        "expect the no-code-credentials error to be shown",
      ).toBeVisible()
      await expect(login.codeInput.input).toBeHidden()

      // The active session is untouched: still aal2, same authentication time.
      const session = await getSession(page.request, kratosPublicURL)
      expect(session.active).toBe(true)
      expect(session.authenticator_assurance_level).toBe("aal2")
      expect(session.authenticated_at).toBe(aal2Session.authenticated_at)
    })
  })
})

// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import {faker} from "@faker-js/faker"
import {expect} from "@playwright/test"
import {test} from "../../../fixtures"
import {LoginPage} from "../../../models/elements/login"
import {getSession, hasNoSession, hasSession} from "../../../actions/session"
import {loginWithPassword} from "../../../actions/login"
import {LoginFlowStyle} from "../../../../shared/config"

const toConfig = ({
                    style = "identifier_first",
                    enumeration = false,
                  }: {
  style?: LoginFlowStyle
  enumeration?: boolean
}) => ({
  selfservice: {
    default_browser_return_url: "http://localhost:4455/welcome",
    flows: {
      login: {
        style,
      },
    },
  },
  security: {
    account_enumeration: {
      mitigate: enumeration,
    },
  },
})

test.use(toConfig({}))

test.describe("password", () => {
  // These can run in parallel because they use the same config.
  test.describe.parallel("account enumeration protection off", () => {
    test.use({
      configOverride: async ({page}, use) => {
        await use(toConfig({style: "identifier_first", enumeration: false}))
      },
    })

    test("login fails because user does not exist", async ({
                                                             page,
                                                             config,
                                                             kratosPublicURL,
                                                           }) => {
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

      const originalSession = await getSession(page.request, kratosPublicURL)
      await login.inputField("password").fill(identity.password)
      await login.submit("password", {
        waitForURL: new RegExp(config.selfservice.default_browser_return_url),
      })
      const newSession = await getSession(page.request, kratosPublicURL)
      expect(originalSession.authenticated_at).not.toEqual(
        newSession.authenticated_at,
      )
    })
  })

  test.describe("account enumeration protection off", () => {
    test.use({
      configOverride: async ({page}, use) => {
        await use(toConfig({style: "identifier_first", enumeration: true}))
      },
    })

    // TODO - write login fails without leaking info
    // TODO write login passes
  })
})
test.describe("oidc", () => {
  test.describe("account enumeration protection off", () => {
    test.use(toConfig({enumeration: false}))

    test("login", async ({page, config, kratosPublicURL}) => {
      await page.goto("/login")

      await page.locator(`button[name=provider][value=hydra]`).click()

      await page
        .locator("input[name=username]")
        .fill(faker.internet.email({provider: "ory.sh"}))
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
})

test.describe("passkeys", () => {
  test.use({
    addVirtualAuthenticator: true,
  })
  test.use({
    configOverride: async ({page}, use) => {
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

  test("login", async ({config, page, kratosPublicURL}) => {
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
  })
})

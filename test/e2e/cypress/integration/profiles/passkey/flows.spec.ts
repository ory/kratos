// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen } from "../../../helpers"
import { routes as express } from "../../../helpers/express"
import { routes as react } from "../../../helpers/react"
import { testFlowWebhook } from "../../../helpers/webhook"

const signup = (registration: string, app: string, email = gen.email()) => {
  cy.visit(registration)

  const websiteTrait = `${
    app === "express" ? `form[data-testid="passkey-flow"]` : ""
  } input[name="traits.website"]`
  const emailTrait = `${
    app === "express" ? `form[data-testid="passkey-flow"]` : ""
  } input[name="traits.email"]`

  cy.get(emailTrait).type(email)
  cy.get(websiteTrait).type("https://www.ory.sh")
  cy.get('[name="passkey_register_trigger"]').click()

  cy.wait(1000)

  cy.getSession({
    expectAal: "aal1",
    expectMethods: ["passkey"],
  }).then((session) => {
    expect(session.identity.traits.email).to.equal(email)
    expect(session.identity.traits.website).to.equal("https://www.ory.sh")
  })
}

context("Passkey registration", () => {
  before(() => {
    cy.task("resetCRI", {})
  })
  after(() => {
    cy.task("resetCRI", {})
  })
  ;[
    {
      login: react.login,
      registration: express.registration,
      settings: react.settings,
      base: react.base,
      app: "react" as "react",
      profile: "passkey",
    },
    {
      login: express.login,
      registration: express.registration,
      settings: express.settings,
      base: express.base,
      app: "express" as "express",
      profile: "passkey",
    },
  ].forEach(({ registration, login, profile, app, base, settings }) => {
    describe(`for app ${app}`, () => {
      let authenticator: any
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)

        cy.task("sendCRI", {
          query: "WebAuthn.enable",
          opts: {},
        })
          .then(() => {
            cy.task("sendCRI", {
              query: "WebAuthn.addVirtualAuthenticator",
              opts: {
                options: {
                  protocol: "ctap2",
                  transport: "internal",
                  hasResidentKey: true,
                  hasUserVerification: true,
                  isUserVerified: true,
                },
              },
            })
          })
          .then((result) => {
            authenticator = result
            cy.log("authenticator ID:", authenticator)
          })

        cy.longPrivilegedSessionTime()
      })

      beforeEach(() => {
        cy.clearAllCookies()
        cy.task("sendCRI", {
          query: "WebAuthn.clearCredentials",
          opts: authenticator,
        })
      })

      after(() => {
        cy.task("sendCRI", {
          query: "WebAuthn.removeVirtualAuthenticator",
          opts: authenticator,
        })
      })

      it("should register after validation errors", () => {
        cy.visit(registration)

        // the browser will prevent the form from being submitted if the input field is required
        // we should remove the required attribute to simulate the data not being sent
        cy.removeAttribute(
          ['input[name="traits.email"]', 'input[name="traits.website"]'],
          "required",
        )

        cy.get(`input[name="traits.website"]`).then(($el) => {
          $el.removeAttr("type")
        })

        const websiteTrait = `${
          app === "express" ? `form[data-testid="passkey-flow"]` : ""
        } input[name="traits.website"]`

        const emailTrait = `${
          app === "express" ? `form[data-testid="passkey-flow"]` : ""
        } input[name="traits.email"]`

        cy.get(websiteTrait).type("b")
        cy.get('[name="passkey_register_trigger"]').click()

        cy.get('[data-testid="ui/message/4000002"]').should("to.exist")
        cy.get('[data-testid="ui/message/4000001"]').should("to.exist")
        cy.get(websiteTrait).should("have.value", "b")

        const email = gen.email()
        cy.get(emailTrait).type(email)
        cy.get('[name="passkey_register_trigger"]').click()

        cy.wait(1000)

        cy.get('[data-testid="ui/message/4000001"]').should("to.exist")
        cy.get(websiteTrait).should("have.value", "b")
        cy.get(emailTrait).should("have.value", email)
        cy.get(websiteTrait).clear()
        cy.get(websiteTrait).type("https://www.ory.sh")

        cy.get('[name="passkey_register_trigger"]').click()

        cy.wait(1000)

        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["passkey"],
        }).then((session) => {
          expect(session.identity.traits.email).to.equal(email)
          expect(session.identity.traits.website).to.equal("https://www.ory.sh")
        })
      })

      it("should pass transient_payload to webhook", () => {
        testFlowWebhook(
          (hooks) =>
            cy.setupHooks("registration", "after", "passkey", [
              ...hooks,
              { hook: "session" },
            ]),
          () => {
            signup(registration, app)
          },
        )
      })

      it("should be able to login with registered account", () => {
        const email = gen.email()

        signup(registration, app, email)
        cy.logout()
        cy.visit(login)

        cy.get('[name="passkey_login_trigger"]').click()
        cy.wait(1000)

        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["passkey"],
        }).then((session) => {
          expect(session.identity.traits.email).to.equal(email)
          expect(session.identity.traits.website).to.equal("https://www.ory.sh")
        })
      })

      it("should not be able to unlink last passkey", () => {
        const email = gen.email()
        signup(registration, app, email)
        cy.visit(settings)
        cy.get('[name="passkey_remove"]').should("have.attr", "disabled")
      })

      it("should be able to link password and use both methods for sign in", () => {
        const email = gen.email()
        const password = gen.password()
        signup(registration, app, email)
        cy.visit(settings)
        cy.get('[name="passkey_remove"]').should("have.attr", "disabled")
        cy.get('[name="password"]').type(password)
        cy.get('[value="password"]').click()
        cy.expectSettingsSaved()
        cy.get('[name="passkey_remove"]').click()
        cy.expectSettingsSaved()
        cy.logout()
        cy.visit(login)

        cy.get('[name="identifier"]').type(email)
        cy.get('[name="password"]').type(password)
        cy.get('[name="method"][value="password"]').click()
      })

      it("should be able to refresh", () => {
        const email = gen.email()
        signup(registration, app, email)
        cy.visit(login + "?refresh=true")
        cy.get('[name="identifier"][type="hidden"]').should("exist")
        cy.get('[name="identifier"][type="input"]').should("not.exist")
        cy.get('[name="password"]').should("not.exist")
        cy.get('[value="password"]').should("not.exist")
        cy.get('[name="passkey_login_trigger"]').click()
        cy.wait(1000)

        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["passkey", "passkey"],
        }).then((session) => {
          expect(session.identity.traits.email).to.equal(email)
          expect(session.identity.traits.website).to.equal("https://www.ory.sh")
        })
      })

      it("should not be able to use for MFA", () => {
        const email = gen.email()
        signup(registration, app, email)
        cy.visit(login + "?aal=aal2")
        cy.get('[value="passkey"]').should("not.exist")
        cy.get('[name="passkey_login_trigger"]').should("not.exist")
      })

      it("should be able to add method later and try a variety of refresh flows", () => {
        const email = gen.email()
        const password = gen.password()
        cy.visit(registration)

        const emailTrait = `${
          app === "express" ? `[data-testid="registration-flow"]` : ""
        } [name="traits.email"]`
        const websiteTrait = `${
          app === "express" ? `[data-testid="registration-flow"]` : ""
        } [name="traits.website"]`

        cy.get(emailTrait).type(email)
        cy.get('[name="password"]').type(password)
        cy.get(websiteTrait).type("https://www.ory.sh")
        cy.get('[value="password"]').click()
        cy.location("pathname").should("not.contain", "/registration")
        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["password"],
        })

        cy.visit(settings)
        cy.get('[name="passkey_register_trigger"]').click()
        cy.expectSettingsSaved()

        cy.visit(login + "?refresh=true")
        cy.get('[name="password"]').should("exist")
        cy.get('[name="passkey_login_trigger"]').click()
        cy.wait(1000)
        cy.location("pathname").should("not.contain", "/login")
        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["password", "passkey", "passkey"],
        })

        cy.visit(login + "?refresh=true")
        cy.get('[name="password"]').type(password)
        cy.get('[value="password"]').click()
        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["password", "passkey", "passkey", "password"],
        })

        cy.logout()
        cy.visit(login)

        cy.get('[name="passkey_login_trigger"]').click()
        cy.wait(1000)
        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["passkey"],
        })
      })
    })
  })
})

// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, gen } from "../../../helpers"
import { routes as express } from "../../../helpers/express"
import { routes as react } from "../../../helpers/react"
import { testRegistrationWebhook } from "../../../helpers/webhook"

const signup = (registration: string, app: string, email = gen.email()) => {
  cy.visit(registration)

  const emailTrait = `${
    app === "express" ? '[data-testid="passwordless-flow"]' : ""
  } [name="traits.email"]`
  const websiteTrait = `${
    app === "express" ? '[data-testid="passwordless-flow"]' : ""
  } [name="traits.website"]`

  cy.get('[name="webauthn_register_displayname"]').type("key1")
  cy.get(emailTrait).type(email)
  cy.get(websiteTrait).type("https://www.ory.sh")
  cy.clickWebAuthButton("register")
  cy.getSession({
    expectAal: "aal1",
    expectMethods: ["webauthn"],
  }).then((session) => {
    expect(session.identity.traits.email).to.equal(email)
    expect(session.identity.traits.website).to.equal("https://www.ory.sh")
  })
}

context("Passwordless registration", () => {
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
      profile: "passwordless",
    },
    {
      login: express.login,
      registration: express.registration,
      settings: express.settings,
      base: express.base,
      app: "express" as "express",
      profile: "passwordless",
    },
  ].forEach(({ registration, login, profile, app, base, settings }) => {
    describe(`for app ${app}`, () => {
      let authenticator
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
        cy.addVirtualAuthenticator().then((result) => {
          authenticator = result
        })
        cy.longPrivilegedSessionTime()
      })

      beforeEach(() => {
        cy.clearAllCookies()
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
          app === "express" ? `[data-testid="passwordless-flow"]` : ""
        } [name="traits.website"]`

        const emailTrait = `${
          app === "express" ? `[data-testid="passwordless-flow"]` : ""
        } [name="traits.email"]`

        cy.get(appPrefix(app) + '[name="webauthn_register_displayname"]').type(
          "key1",
        )
        cy.get(websiteTrait).type("b")
        cy.clickWebAuthButton("register")

        cy.get('[data-testid="ui/message/4000002"]').should("to.exist")
        cy.get('[data-testid="ui/message/4000001"]').should("to.exist")
        cy.get(websiteTrait).should("have.value", "b")

        const email = gen.email()
        cy.get(emailTrait).type(email)
        cy.clickWebAuthButton("register")

        cy.get('[data-testid="ui/message/4000001"]').should("to.exist")
        cy.get(websiteTrait).should("have.value", "b")
        cy.get(emailTrait).should("have.value", email)
        cy.get(websiteTrait).clear().type("https://www.ory.sh")
        cy.clickWebAuthButton("register")
        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["webauthn"],
        }).then((session) => {
          expect(session.identity.traits.email).to.equal(email)
          expect(session.identity.traits.website).to.equal("https://www.ory.sh")
        })
      })

      it("should pass transient_payload to webhook", () => {
        testRegistrationWebhook(
          (hooks) => cy.setupHooks("registration", "after", "webauthn", hooks),
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

        const identifierTrait = `${
          app === "express" ? `[data-testid="passwordless-flow"]` : ""
        } [name="identifier"]`

        cy.get(identifierTrait).type(email)
        cy.get('[value="webauthn"]').click()
        cy.get('[data-testid="ui/message/1010012"]').should("to.exist")
        cy.get('[name="password"]').should("to.not.exist")
        cy.clickWebAuthButton("login")
        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["webauthn"],
        }).then((session) => {
          expect(session.identity.traits.email).to.equal(email)
          expect(session.identity.traits.website).to.equal("https://www.ory.sh")
        })
      })

      it("should not be able to unlink last security key", () => {
        const email = gen.email()
        signup(registration, app, email)
        cy.visit(settings)
        cy.get('[name="webauthn_remove"]').should("not.exist")
      })

      it("should be able to link password and use both methods for sign in", () => {
        const email = gen.email()
        const password = gen.password()
        signup(registration, app, email)
        cy.visit(settings)
        cy.get('[name="webauthn_remove"]').should("not.exist")
        cy.get('[name="password"]').type(password)
        cy.get('[value="password"]').click()
        cy.expectSettingsSaved()
        cy.get('[name="webauthn_remove"]').click()
        cy.expectSettingsSaved()
        cy.logout()
        cy.visit(login)

        const identifierTrait = `${
          app === "express" ? `[data-testid="passwordless-flow"]` : ""
        } [name="identifier"]`

        cy.get(identifierTrait).type(email)
        cy.get('[value="webauthn"]').click()
        cy.get('[data-testid="ui/message/4000015"]').should("to.exist")
        cy.get(identifierTrait).should("exist")
        cy.get('[name="password"]').should("exist")
        cy.get('[value="password"]').should("exist")
      })

      it("should be able to refresh", () => {
        const email = gen.email()
        signup(registration, app, email)
        cy.visit(login + "?refresh=true")
        cy.get('[name="identifier"][type="hidden"]').should("exist")
        cy.get('[name="identifier"][type="input"]').should("not.exist")
        cy.get('[name="password"]').should("not.exist")
        cy.get('[value="password"]').should("not.exist")
        cy.clickWebAuthButton("login")
        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["webauthn", "webauthn"],
        }).then((session) => {
          expect(session.identity.traits.email).to.equal(email)
          expect(session.identity.traits.website).to.equal("https://www.ory.sh")
        })
      })

      it("should not be able to use for MFA", () => {
        const email = gen.email()
        signup(registration, app, email)
        cy.visit(login + "?aal=aal2")
        cy.get('[value="webauthn"]').should("not.exist")
        cy.get('[name="webauthn_login_trigger"]').should("not.exist")
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
        cy.get('[name="webauthn_register_displayname"]').type("key2")
        cy.clickWebAuthButton("register")
        cy.expectSettingsSaved()

        cy.visit(login + "?refresh=true")
        cy.get('[name="password"]').should("exist")
        cy.clickWebAuthButton("login")
        cy.location("pathname").should("not.contain", "/login")
        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["password", "webauthn", "webauthn"],
        })

        cy.visit(login + "?refresh=true")
        cy.get('[name="password"]').type(password)
        cy.get('[value="password"]').click()
        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["password", "webauthn", "webauthn", "password"],
        })

        cy.logout()
        cy.visit(login)

        const identifierTrait = `${
          app === "express" ? `[data-testid="passwordless-flow"]` : ""
        } [name="identifier"]`

        cy.get(identifierTrait).type(email)
        cy.get('[value="webauthn"]').click()
        cy.clickWebAuthButton("login")
        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["webauthn"],
        })
      })

      it("should not be able to use for MFA even when passwordless is false", () => {
        const email = gen.email()
        signup(registration, app, email)
        cy.updateConfigFile((config) => {
          config.selfservice.methods.webauthn.config.passwordless = false
          return config
        })
        cy.visit(login + "?aal=aal2")
        cy.get('[value="webauthn"]').should("not.exist")
        cy.get('[name="webauthn_login_trigger"]').should("not.exist")

        cy.visit(settings)
        cy.get('[name="webauthn_remove"]').should("not.exist")
        cy.get('[name="webauthn_register_displayname"]').type("key2")
        cy.clickWebAuthButton("register")
        cy.expectSettingsSaved()

        cy.visit(login + "?aal=aal2&refresh=true")
        cy.clickWebAuthButton("login")
        cy.getSession({
          expectAal: "aal2",
          expectMethods: ["webauthn", "webauthn", "webauthn"],
        })
      })
    })
  })
})

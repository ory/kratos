// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, gen, website } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"
import { testRegistrationWebhook } from "../../../../helpers/webhook"

context("Social Sign Up Successes", () => {
  ;[
    {
      login: react.login,
      registration: react.registration,
      app: "react" as "react",
      profile: "spa",
    },
    {
      login: express.login,
      registration: express.registration,
      app: "express" as "express",
      profile: "oidc",
    },
  ].forEach(({ registration, login, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.clearAllCookies()
        cy.visit(registration)
        cy.setIdentitySchema(
          "file://test/e2e/profiles/oidc/identity.traits.schema.json",
        )
      })

      const shouldSession = (email) => (session) => {
        const { identity } = session
        expect(identity.id).to.not.be.empty
        expect(identity.traits.website).to.equal(website)
        expect(identity.traits.email).to.equal(email)
      }

      it("should be able to sign up with incomplete data and finally be signed in", () => {
        const email = gen.email()

        cy.registerOidc({
          app,
          email,
          expectSession: false,
          route: registration,
        })

        cy.get("#registration-password").should("not.exist")
        cy.get(appPrefix(app) + '[name="traits.email"]').should(
          "have.value",
          email,
        )
        cy.get('[data-testid="ui/message/4000002"]').should(
          "contain.text",
          "Property website is missing",
        )

        cy.get('[name="traits.consent"][type="checkbox"]')
          .siblings("label")
          .click()
        cy.get('[name="traits.newsletter"][type="checkbox"]')
          .siblings("label")
          .click()
        cy.get('[name="traits.website"]').type("http://s")

        cy.get('[name="provider"]')
          .should("have.length", 1)
          .should("have.value", "hydra")
          .should("contain.text", "Continue")
          .click()

        cy.get("#registration-password").should("not.exist")
        cy.get('[name="traits.email"]').should("have.value", email)
        cy.get('[name="traits.website"]').should("have.value", "http://s")
        cy.get('[data-testid="ui/message/4000003"]').should(
          "contain.text",
          "length must be >= 10",
        )
        cy.get('[name="traits.website"]')
          .should("have.value", "http://s")
          .clear()
          .type(website)

        cy.get('[name="traits.consent"]').should("be.checked")
        cy.get('[name="traits.newsletter"]').should("be.checked")

        cy.triggerOidc(app)

        cy.location("pathname").should((loc) => {
          expect(loc).to.be.oneOf(["/welcome", "/", "/sessions"])
        })

        cy.getSession().should((session) => {
          shouldSession(email)(session)
          expect(session.identity.traits.consent).to.equal(true)
        })
      })

      it("should pass transient_payload to webhook", () => {
        testRegistrationWebhook(
          (hooks) => cy.setupHooks("registration", "after", "oidc", hooks),
          () => {
            const email = gen.email()
            cy.registerOidc({
              app,
              email,
              website,
              route: registration,
            })
            cy.getSession().should(shouldSession(email))
          },
        )
      })

      it("should be able to sign up with complete data", () => {
        const email = gen.email()

        cy.registerOidc({ app, email, website, route: registration })
        cy.getSession().should(shouldSession(email))
      })

      it("should be able to convert a sign up flow to a sign in flow", () => {
        const email = gen.email()

        cy.registerOidc({ app, email, website, route: registration })
        cy.logout()
        cy.noSession()
        cy.visit(registration)
        cy.triggerOidc(app)

        cy.location("pathname").should((path) => {
          expect(path).to.oneOf(["/", "/welcome", "/sessions"])
        })

        cy.getSession().should(shouldSession(email))
      })

      it("should be able to convert a sign in flow to a sign up flow", () => {
        cy.setIdentitySchema(
          "file://test/e2e/profiles/oidc/identity-required.traits.schema.json",
        )

        const email = gen.email()
        cy.visit(login)
        cy.triggerOidc(app)

        cy.get("#username").clear().type(email)
        cy.get("#remember").click()
        cy.get("#accept").click()
        cy.get('[name="scope"]').each(($el) => cy.wrap($el).click())
        cy.get("#remember").click()
        cy.get("#accept").click()

        cy.get('[data-testid="ui/message/4000002"]').should(
          "contain.text",
          "Property website is missing",
        )
        cy.get('[name="traits.website"]').type("http://s")

        cy.triggerOidc(app)

        cy.get('[data-testid="ui/message/4000003"]').should(
          "contain.text",
          "length must be >= 10",
        )
        cy.get('[name="traits.requirednested"]').should("not.exist")
        cy.get('[name="traits.requirednested.a"]').siblings("label").click()
        cy.get('[name="traits.consent"]').siblings("label").click()
        cy.get('[name="traits.website"]')
          .should("have.value", "http://s")
          .clear()
          .type(website)
        cy.triggerOidc(app)

        cy.location("pathname").should("not.contain", "/registration")

        cy.getSession().should(shouldSession(email))
      })

      it("should be able to sign up with redirects", () => {
        const email = gen.email()
        cy.registerOidc({
          app,
          email,
          website,
          route: registration + "?return_to=https://www.ory.sh/",
        })
        cy.location("href").should("eq", "https://www.ory.sh/")
        cy.logout()
      })

      it("should be able to register with upstream parameters", () => {
        const email = gen.email()
        cy.intercept("GET", "**/oauth2/auth*").as("getHydraRegistration")

        cy.visit(registration + "?return_to=https://www.example.org/")

        cy.addInputElement("form", "upstream_parameters.login_hint", email)

        cy.triggerOidc(app)

        // once a request to getHydraRegistration responds, 'cy.wait' will resolve
        cy.wait("@getHydraRegistration")
          .its("request.url")
          .should("include", "login_hint=" + encodeURIComponent(email))
      })

      it("oidc registration with duplicate identifier should return new login flow with duplicate error", () => {
        cy.visit(registration)

        const email = gen.email()
        const password = gen.password()

        cy.get('input[name="traits.email"]').type(email)
        cy.get('input[name="password"]').type(password)
        cy.get('input[name="traits.website"]').type(website)
        cy.get('[name="traits.consent"]').siblings("label").click()
        cy.get('[name="traits.newsletter"]').siblings("label").click()

        cy.submitPasswordForm()

        cy.location("pathname").should("not.contain", "/registration")

        cy.getSession().should(shouldSession(email))

        cy.logout()
        cy.noSession()

        // register the account through the OIDC provider
        cy.registerOidc({
          app,
          acceptConsent: true,
          acceptLogin: true,
          expectSession: false,
          email,
          website,
          route: registration,
        })

        cy.get('[data-testid="ui/message/4000027"]').should("be.visible")

        cy.location("href").should("contain", "/login")

        cy.get("[name='provider'][value='hydra']").should("be.visible")
        cy.get("[name='provider'][value='google']").should("be.visible")
        cy.get("[name='provider'][value='github']").should("be.visible")

        if (app === "express") {
          cy.get("[data-testid='forgot-password-link']").should("be.visible")
        }

        cy.get("input[name='identifier']").type(email)
        cy.get("input[name='password']").type(password)
        cy.submitPasswordForm()
        cy.getSession()
      })
    })
  })
})

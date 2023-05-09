// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, APP_URL, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Registration success with email profile", () => {
  ;[
    {
      route: express.registration,
      app: "express" as "express",
      profile: "email",
    },
    {
      route: react.registration,
      app: "react" as "react",
      profile: "spa",
    },
  ].forEach(({ route, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.deleteMail()
        cy.clearAllCookies()
        cy.visit(route)
        cy.enableVerification()
        if (app === "express") {
          cy.enableVerificationUIAfterRegistration("password")
        }
      })

      it("should sign up and be logged in", () => {
        const email = gen.email()
        const password = gen.password()
        const website = "https://www.example.org/"
        const age = 30

        cy.get(appPrefix(app) + 'input[name="traits"]').should("not.exist")
        cy.get('input[name="traits.email"]').type(email)
        cy.get('input[name="password"]').type(password)
        cy.get('input[name="traits.website').type(website)
        cy.get('input[name="traits.age"]').type(`${age}`)
        cy.get('[type="checkbox"][name="traits.tos"]').click({ force: true })

        cy.submitPasswordForm()

        cy.url().should("contain", "verification")
        cy.getVerificationCodeFromEmail(email).then((code) => {
          cy.get("input[name=code]").type(code)
          cy.get("button[name=method][value=code]").click()
        })

        cy.get('[data-testid="ui/message/1080002"]').should(
          "have.text",
          "You successfully verified your email address.",
        )

        cy.get("[data-testid='node/anchor/continue']").click()

        if (app === "express") {
          cy.get('a[href*="sessions"').click()
        }
        cy.get("pre").should("contain.text", email)

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.verifiable_addresses).to.have.length(1)
          expect(identity.schema_id).to.equal("default")
          expect(identity.schema_url).to.equal(`${APP_URL}/schemas/ZGVmYXVsdA`)
          expect(identity.traits.website).to.equal(website)
          expect(identity.traits.email).to.equal(email)
          expect(identity.traits.age).to.equal(age)
          expect(identity.traits.tos).to.equal(true)
        })
      })

      it("should sign up with advanced form field values be logged in", () => {
        const email = gen.email()
        const password = gen.password()

        cy.get('input[name="traits"]').should("not.exist")
        cy.get('input[name="traits.email"]').type(email)
        cy.get('input[name="password"]').type(password)
        const website = "https://www.example.org/"
        cy.get('input[name="traits.website"]').type(website)

        cy.submitPasswordForm()

        cy.url().should("contain", "verification")
        cy.getVerificationCodeFromEmail(email).then((code) => {
          cy.get("input[name=code]").type(code)
          cy.get("button[name=method][value=code]").click()
        })

        cy.get('[data-testid="ui/message/1080002"]').should(
          "have.text",
          "You successfully verified your email address.",
        )

        cy.get("[data-testid='node/anchor/continue']").click()

        if (app === "express") {
          cy.get('a[href*="sessions"').click()
        }
        cy.get("pre").should("contain.text", email)

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.verifiable_addresses).to.have.length(1)
          expect(identity.schema_id).to.equal("default")
          expect(identity.schema_url).to.equal(`${APP_URL}/schemas/ZGVmYXVsdA`)
          expect(identity.traits.website).to.equal(website)
          expect(identity.traits.email).to.equal(email)
          expect(identity.traits.age).to.be.undefined
          expect(identity.traits.tos).to.be.oneOf([false, undefined])
        })
      })

      it("should sign up and be redirected", () => {
        cy.disableVerification()
        cy.browserReturnUrlOry()
        cy.visit(route + "?return_to=https://www.example.org/")

        const email = gen.email()
        const password = gen.password()
        const website = "https://www.example.org/"

        cy.get('input[name="traits"]').should("not.exist")
        cy.get('input[name="traits.email"]').type(email)
        cy.get('input[name="traits.website').type(website)
        cy.get('input[name="password"]').type(password)
        cy.submitPasswordForm()

        cy.url().should("eq", "https://www.example.org/")
      })
    })
  })

  describe("redirect for express app", () => {
    it("should redirect to return_to after flow expires", () => {
      // Wait for flow to expire
      cy.useConfigProfile("email")
      cy.shortRegisterLifespan()
      cy.browserReturnUrlOry()
      cy.proxy("express")
      cy.visit(express.registration + "?return_to=https://www.example.org/")
      cy.wait(105)

      const email = gen.email()
      const password = gen.password()
      const website = "https://www.example.org/"

      cy.get(`${appPrefix("express")} input[name="traits"]`).should("not.exist")
      cy.get('input[name="traits.email"]').type(email)
      cy.get('input[name="traits.website').type(website)
      cy.get('input[name="password"]').type(password)

      cy.longRegisterLifespan()
      cy.submitPasswordForm()

      cy.get('[data-testid="ui/message/4040001"]').should(
        "contain.text",
        "The registration flow expired",
      )

      // Try again with long lifespan set
      cy.get('input[name="traits"]').should("not.exist")
      cy.get('input[name="traits.email"]').type(email)
      cy.get('input[name="traits.website').type(website)
      cy.get('input[name="password"]').type(password)
      cy.submitPasswordForm()

      cy.url().should("eq", "https://www.example.org/")
    })

    it("should not redirect to verification_flow if not configured", () => {
      cy.deleteMail()
      cy.useConfigProfile("email")
      cy.enableVerification()
      cy.proxy("express")
      cy.visit(express.registration + "?return_to=https://www.example.org/")

      const email = gen.email()
      const password = gen.password()
      const website = "https://www.example.org/"

      cy.get(`${appPrefix("express")} input[name="traits"]`).should("not.exist")
      cy.get('input[name="traits.email"]').type(email)
      cy.get('input[name="traits.website').type(website)
      cy.get('input[name="password"]').type(password)

      cy.submitPasswordForm()

      // Verify that the verification code is still sent
      cy.getVerificationCodeFromEmail(email).should("not.be.undefined")

      cy.url().should("eq", "https://www.example.org/")
    })
  })
})

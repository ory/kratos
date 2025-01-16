// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, APP_URL, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Registration success with two-step signup", () => {
  ;[
    {
      route: express.registration,
      app: "express" as "express",
      profile: "two-steps",
    },
    {
      route: react.registration,
      app: "react" as "react",
      profile: "two-steps",
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

        // Fill out step one forms
        cy.get(appPrefix(app) + 'input[name="traits"]').should("not.exist")
        cy.get('input[name="traits.email"]').type("someone@foo")
        cy.get('input[name="traits.website"]').type(website)
        cy.get('[name="method"][value="profile"]').click()

        // navigate back, fill traits again
        cy.get('[name="screen"][value="previous"]').click()
        cy.get('input[name="traits.email"]').type(
          "{selectall}{backspace}" + email,
        )
        cy.get('[name="method"][value="profile"]').click()

        // Fill out step two forms
        cy.get('input[name="password"]').type(password)
        cy.get('[name="method"][value="password"]').click()

        if (app === "express") {
          cy.get('a[href*="sessions"]').click()
        }
        cy.get("pre").should("contain.text", email)

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.schema_id).to.equal("default")
          expect(identity.schema_url).to.equal(`${APP_URL}/schemas/ZGVmYXVsdA`)
          expect(identity.traits.website).to.equal(website)
          expect(identity.traits.email).to.equal(email)
        })
      })

      it("should handle form errors", () => {
        const email = gen.email()
        const password = gen.password()
        const websiteTooShort = "a://b"
        const website = "https://www.example.com"

        // Fill out step one forms
        cy.get(appPrefix(app) + 'input[name="traits"]').should("not.exist")
        cy.get('input[name="traits.email"]').type(email)
        cy.get('input[name="traits.website"]').type(websiteTooShort)
        cy.get('[name="method"][value="profile"]').click()

        // Assert form errors
        cy.get('[data-testid="ui/message/4000003"]').should("to.exist")

        // Enter correct value
        cy.get('input[name="traits.website"]').type(
          "{selectall}{backspace}" + website,
        )
        cy.get('[name="method"][value="profile"]').click()

        // Fill out step two forms
        cy.get('input[name="password"]').type(password)
        cy.get('[name="method"][value="password"]').click()

        if (app === "express") {
          cy.get('a[href*="sessions"]').click()
        }
        cy.get("pre").should("contain.text", email)

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.schema_id).to.equal("default")
          expect(identity.schema_url).to.equal(`${APP_URL}/schemas/ZGVmYXVsdA`)
          expect(identity.traits.website).to.equal(website)
          expect(identity.traits.email).to.equal(email)
        })
      })
    })
  })
})

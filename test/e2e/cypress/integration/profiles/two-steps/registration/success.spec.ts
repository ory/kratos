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
        const age = 30

        // Fill out step one forms
        cy.get(appPrefix(app) + 'input[name="traits"]').should("not.exist")
        cy.get('input[name="traits.email"]').type(email)
        cy.get('input[name="traits.website').type(website)
        cy.get('input[name="traits.age"]').type(`${age}`)
        cy.get('[type="checkbox"][name="traits.tos"]').click({ force: true })
        cy.get('[name="method"][value="profile"]').click()

        // Fill out step two forms
        cy.get('input[name="password"]').type(password)
        cy.get('[name="method"][value="password"]').click()

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
    })
  })
})

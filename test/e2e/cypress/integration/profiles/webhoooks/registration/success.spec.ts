// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { APP_URL, appPrefix, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"

context("Registration success with email profile with webhooks", () => {
  ;[
    {
      route: express.registration,
      app: "express" as "express",
      profile: "webhooks",
    },
  ].forEach(({ route, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.clearAllCookies()
        cy.visit(route)
      })

      it("should sign up and be logged in", () => {
        const email = gen.email()
        const password = gen.password()

        cy.get(appPrefix(app) + 'input[name="traits"]').should("not.exist")
        cy.get('input[name="traits.email"]').type(email)
        cy.get('input[name="password"]').type(password)

        cy.submitPasswordForm()
        if (app === "express") {
          cy.get("a[href*='sessions']").click()
        }
        cy.get("pre").should("contain.text", email)

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.schema_id).to.equal("default")
          expect(identity.schema_url).to.equal(`${APP_URL}/schemas/ZGVmYXVsdA`)
          expect(identity.traits.email).to.equal(email)
        })
      })
    })
  })
})

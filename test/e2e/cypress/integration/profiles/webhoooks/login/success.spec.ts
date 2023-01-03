// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { APP_URL, appPrefix, gen, website } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"

describe("Basic email profile with succeeding login flows with webhooks", () => {
  const email = gen.email()
  const password = gen.password()

  before(() => {
    cy.registerApi({ email, password, fields: { "traits.website": website } })
  })
  ;[
    {
      route: express.login,
      app: "express" as "express",
      profile: "webhooks",
    },
  ].forEach(({ route, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.useConfigProfile(profile)
        cy.clearAllCookies()
        cy.visit(route)
      })

      it("should sign in and be logged in", () => {
        cy.get(`${appPrefix(app)}input[name="identifier"]`).type(email)
        cy.get('input[name="password"]').type(password)
        cy.submitPasswordForm()
        cy.location("pathname").should("not.contain", "/login")

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

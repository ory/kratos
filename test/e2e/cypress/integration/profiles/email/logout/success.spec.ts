// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, gen, website } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Testing logout flows", () => {
  ;[
    {
      route: express.login,
      app: "express" as "express",
      profile: "email",
    },
    {
      route: react.login,
      app: "react" as "react",
      profile: "spa",
    },
  ].forEach(({ route, profile, app }) => {
    describe(`for app ${app}`, () => {
      const email = gen.email()
      const password = gen.password()

      before(() => {
        cy.proxy(app)

        cy.useConfigProfile(profile)
        cy.registerApi({
          email,
          password,
          fields: { "traits.website": website },
        })
      })

      beforeEach(() => {
        cy.clearAllCookies()
        cy.login({ email, password, cookieUrl: route })
        cy.visit(route)
      })

      it("should sign out and be able to sign in again", () => {
        cy.getSession()
        cy.getCookie("ory_kratos_session").should("not.be.null")
        if (app === "express") {
          cy.get(
            `${appPrefix(app)} [data-testid="logout"] a:not(.disabled)`,
          ).click()
        } else {
          cy.get(
            `${appPrefix(app)} [data-testid="logout"]:not(.disabled)`,
          ).click()
        }
        cy.getCookie("ory_kratos_session").should("be.null")
        cy.noSession()
        cy.url().should("include", "/login")
      })
    })
  })
})

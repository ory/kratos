// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"
import { appPrefix } from "../../../../helpers"

context("UI tests using the email profile", () => {
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
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.visit(route)
      })

      it("should use the json schema titles", () => {
        cy.get(`${appPrefix(app)}input[name="identifier"]`)
          .parent()
          .should("contain.text", "ID")
        cy.get('input[name="password"]')
          .parent()
          .should("contain.text", "Password")
        cy.get('button[value="password"]').should("contain.text", "Sign in")
      })

      it("clicks the log in link", () => {
        cy.get('a[href*="registration"]').click()
        cy.location("pathname").should("include", "registration")

        if (app === "express") {
          cy.location("search").should("not.be.empty")
        }
      })
    })
  })
})

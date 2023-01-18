// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Settings errors with email profile", () => {
  ;[
    {
      base: express.base,
      app: "express" as "express",
      profile: "email",
    },
    {
      base: react.base,
      app: "react" as "react",
      profile: "spa",
    },
  ].forEach(({ profile, app, base }) => {
    describe(`for app ${app}`, () => {
      const identity = gen.identity()

      before(() => {
        cy.useConfigProfile(profile)
        cy.registerApi({
          ...identity,
          fields: { "traits.website": "https://www.ory.sh/" },
        })
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.login({ ...identity, cookieUrl: base })
        cy.visit(base)
      })

      describe("use ui elements", () => {
        it("should use the json schema titles", () => {
          const settingsLink = appPrefix(app) + 'a[href*="settings"]'
          if (app === "express") {
            cy.get(settingsLink).should("have.attr", "target", "_blank")
            cy.removeAttribute([settingsLink], "target")
          }
          cy.get(settingsLink).click()
          cy.get('input[name="traits.email"]')
            .parent()
            .should("contain.text", "Your E-Mail")
          cy.get('input[name="traits.website"]')
            .parent()
            .should("contain.text", "Your website")
          cy.get('input[name="password"]')
            .parent()
            .should("contain.text", "Password")
          cy.get('button[value="profile"]').should("contain.text", "Save")
          cy.get('button[value="password"]').should("contain.text", "Save")
        })

        it("clicks the settings link", () => {
          const settingsLink = 'a[href*="settings"]'
          if (app === "express") {
            cy.get(settingsLink).should("have.attr", "target", "_blank")
            cy.removeAttribute([settingsLink], "target")
          }
          cy.get(settingsLink).click()
          cy.location("pathname").should("include", "settings")
        })
      })
    })
  })
})

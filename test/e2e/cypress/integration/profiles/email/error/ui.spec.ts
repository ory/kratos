// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"
import { appPrefix } from "../../../../helpers"

describe("Handling self-service error flows", () => {
  ;[
    {
      route: express.base,
      app: "express" as "express",
      profile: "email",
    },
    {
      route: react.base,
      app: "react" as "react",
      profile: "spa",
    },
  ].forEach(({ route, app, profile }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      it("should show the error", () => {
        cy.visit(`${route}/error?id=stub:500`, {
          failOnStatusCode: false,
        })

        if (app === "express") {
          cy.get(`${appPrefix(app)} [data-testid="ui/error/message"]`).should(
            "contain.text",
            "This is a stub error.",
          )
        } else {
          cy.get(`${appPrefix(app)}code`).should(
            "contain.text",
            "This is a stub error.",
          )
        }
      })
    })
  })
})

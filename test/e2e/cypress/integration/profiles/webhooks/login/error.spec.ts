// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"

describe("Basic email profile with failing login flows with webhooks", () => {
  ;[
    {
      route: express.login,
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

      it("should show fail to sign in when webhooks rejects login", () => {
        const email = gen.blockedEmail()
        const password = gen.password()

        cy.registerApi({ email, password, fields: {} })
        cy.get('input[name="identifier"]').type(email)
        cy.get('input[name="password"]').type(password)

        cy.submitPasswordForm()
        cy.get('*[data-testid="ui/message/1234"]').should(
          "contain.text",
          "email could not be validated",
        )
      })
    })
  })
})

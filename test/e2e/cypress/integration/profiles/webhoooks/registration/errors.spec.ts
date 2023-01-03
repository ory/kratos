// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"

describe("Registration failures with email profile with webhooks", () => {
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
        cy.visit(route)
      })

      const blockedIdentity = gen.blockedEmail()
      const password = gen.password()

      it("should show an error when the webhook is blocking registration", () => {
        cy.get('input[name="traits.email"]').type(blockedIdentity)
        cy.get('input[name="password"]').type(password)

        cy.submitPasswordForm()
        cy.get('input[name="traits.email"]').should(
          "have.value",
          blockedIdentity,
        )
        cy.get('*[data-testid="ui/message/1234"]').should(
          "contain.text",
          "email could not be validated",
        )
      })
    })
  })
})

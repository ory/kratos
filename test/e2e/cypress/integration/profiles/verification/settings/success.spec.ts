// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import {
  APP_URL,
  appPrefix,
  assertVerifiableAddress,
  gen,
} from "../../../../helpers"
import { routes as react } from "../../../../helpers/react"
import { routes as express } from "../../../../helpers/express"

context("Account Verification Settings Success", () => {
  ;[
    {
      settings: react.settings,
      app: "react" as "react",
      profile: "verification",
    },
    {
      settings: express.settings,
      app: "express" as "express",
      profile: "verification",
    },
  ].forEach(({ profile, settings, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })
      let identity

      before(() => {
        cy.deleteMail()
      })

      beforeEach(() => {
        identity = gen.identity()
        cy.register(identity)
        cy.deleteMail({ atLeast: 1 }) // clean up registration email

        cy.login(identity)
        cy.visit(settings)
      })

      it("should update the verify address and request a verification email", () => {
        const email = `not-${identity.email}`
        cy.get(appPrefix(app) + 'input[name="traits.email"]')
          .clear()
          .type(email)
        cy.get('[value="profile"]').click()
        cy.expectSettingsSaved()
        cy.get('input[name="traits.email"]').should("contain.value", email)
        cy.getSession().then(
          assertVerifiableAddress({ isVerified: false, email }),
        )

        cy.verifyEmail({ expect: { email } })
      })

      xit("should should be able to allow or deny (and revert?) the address change", () => {
        // FIXME https://github.com/ory/kratos/issues/292
      })
    })
  })
})

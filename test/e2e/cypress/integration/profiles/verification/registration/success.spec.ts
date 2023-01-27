// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { assertVerifiableAddress, gen } from "../../../../helpers"

import { routes as react } from "../../../../helpers/react"
import { routes as express } from "../../../../helpers/express"

context("Account Verification Registration Success", () => {
  ;[
    {
      registration: react.registration,
      app: "react" as "react",
      profile: "verification",
    },
    {
      registration: express.registration,
      app: "express" as "express",
      profile: "verification",
    },
  ].forEach(({ profile, registration, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.longVerificationLifespan()
        cy.deleteMail()
      })

      afterEach(() => {
        cy.deleteMail()
      })

      const up = (value) => `up-${value}`
      const { email, password } = gen.identity()

      it("is able to verify the email address after sign up", () => {
        const identity = gen.identityWithWebsite()
        const { email, password } = identity
        cy.registerApi(identity)
        cy.login(identity)
        cy.getSession().should((session) =>
          assertVerifiableAddress({
            isVerified: false,
            email,
          })(session),
        )

        cy.verifyEmail({ expect: { email, password } })
      })

      xit("sends the warning email on double sign up", () => {
        // FIXME https://github.com/ory/kratos/issues/133
        cy.clearAllCookies()
        cy.register({ email, password: up(password) })
        cy.clearAllCookies()
        cy.login({ email, password })

        cy.verifyEmail({ expect: { email, password } })
      })

      it("is redirected to after_verification_return_to after verification", () => {
        cy.clearAllCookies()
        const { email, password } = gen.identity()
        cy.register({
          email,
          password,
          query: {
            after_verification_return_to:
              "http://localhost:4455/verification_callback",
          },
        })
        cy.login({ email, password })
        cy.verifyEmail({
          expect: {
            email,
            password,
            redirectTo: "http://localhost:4455/verification_callback",
          },
        })
      })
    })
  })
})

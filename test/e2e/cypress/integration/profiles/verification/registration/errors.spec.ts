// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import {
  appPrefix,
  assertVerifiableAddress,
  gen,
  parseHtml,
  verifyHrefPattern,
} from "../../../../helpers"

import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Account Verification Registration Errors", () => {
  ;[
    {
      login: react.login,
      app: "react" as "react",
      profile: "verification",
    },
    {
      login: express.login,
      app: "express" as "express",
      profile: "verification",
    },
  ].forEach(({ profile, login, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      let identity
      beforeEach(() => {
        cy.enableVerification()
        cy.disableRecovery()
        cy.shortCodeLifespan()
        cy.longVerificationLifespan()

        cy.deleteMail()

        identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.login(identity)
      })

      it("is unable to verify the email address if the code is no longer valid and resend the code", () => {
        cy.shortCodeLifespan()
        cy.verifyEmailButExpired({
          expect: { email: identity.email },
        })

        cy.longCodeLifespan()

        cy.get(appPrefix(app) + 'input[name="email"]').should("be.empty")
        cy.get('input[name="email"]').type(identity.email)
        cy.get('button[value="code"]').click()
        cy.contains("An email containing a verification")
        cy.verifyEmail({
          expect: { email: identity.email, password: identity.password },
        })
      })

      it("is unable to verify the email address if the code is incorrect", () => {
        cy.getMail().then((mail) => {
          const link = parseHtml(mail.body).querySelector("a")

          expect(verifyHrefPattern.test(link.href)).to.be.true

          cy.visit(link.href + "-not") // add random stuff to the confirm challenge
          cy.getSession().should((session) =>
            assertVerifiableAddress({
              isVerified: false,
              email: identity.email,
            })(session),
          )
        })
      })
    })
  })
})

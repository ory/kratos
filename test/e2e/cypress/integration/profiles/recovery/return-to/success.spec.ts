// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { authenticator } from "otplib"
import { appPrefix, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Recovery with `return_to`", () => {
  ;[
    {
      recovery: react.recovery,
      settings: react.settings,
      base: react.base,
      app: "react" as "react",
      profile: "recovery-mfa",
    },
    {
      recovery: express.recovery,
      settings: express.settings,
      app: "express" as "express",
      profile: "recovery-mfa",
    },
  ].forEach(({ app, recovery, settings, profile }) => {
    describe("Recovery with `return_to` query paramter success", () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      let identity: any

      beforeEach(() => {
        cy.deleteMail()
        cy.longRecoveryLifespan()
        cy.disableVerification()
        cy.enableRecovery()
        cy.useRecoveryStrategy("code")
        cy.notifyUnknownRecipients("recovery", false)
        cy.clearAllCookies()
        cy.longPrivilegedSessionTime()
        cy.requireStrictAal()
        identity = gen.identityWithWebsite()
        cy.registerApi(identity)
      })

      const doRecovery = () => {
        cy.get(appPrefix(app) + "input[name='email']").type(identity.email)
        cy.get("button[value='code']").click()

        cy.get('[data-testid="ui/message/1060003"]').should(
          "have.text",
          "An email containing a recovery code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
        )

        cy.recoveryEmailWithCode({ expect: { email: identity.email } })
        cy.get("button[value='code']").click()
      }

      it("should return to the `return_to` url after successful account recovery and settings update", () => {
        cy.visit(recovery + "?return_to=https://www.ory.sh/")
        doRecovery()

        cy.get('[data-testid="ui/message/1060001"]', { timeout: 30000 }).should(
          "contain.text",
          "You successfully recovered your account. ",
        )

        cy.getSession()
        cy.location("pathname").should("eq", "/settings")

        const newPassword = gen.password()
        cy.get(appPrefix(app) + 'input[name="password"]')
          .clear()
          .type(newPassword)
        cy.get('button[value="password"]').click()

        cy.location("hostname").should("eq", "www.ory.sh")
      })

      it("should return to the `return_to` url even with mfa enabled after successful account recovery and settings update", () => {
        cy.requireStrictAal()

        cy.visit(settings)
        cy.get('input[name="identifier"]').type(identity.email)
        cy.get('input[name="password"]').type(identity.password)
        cy.get('button[value="password"]').click()
        cy.visit(settings)

        // enable mfa for this account
        let secret: string
        cy.get('[data-testid="node/text/totp_secret_key/text"]').then(($e) => {
          secret = $e.text().trim()
        })
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[name="method"][value="totp"]').click()
        cy.expectSettingsSaved()
        cy.getSession({
          expectAal: "aal2",
          expectMethods: ["password", "totp"],
        })

        cy.logout()
        cy.clearAllCookies()

        cy.visit(recovery + "?return_to=https://www.ory.sh/")
        doRecovery()

        cy.shouldShow2FAScreen()
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[name="method"][value="totp"]').click()

        const newPassword = gen.password()
        cy.get(appPrefix(app) + 'input[name="password"]')
          .clear()
          .type(newPassword)
        cy.get('button[value="password"]').click()
        cy.location("hostname").should("eq", "www.ory.sh")
      })
    })
  })
})

// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, assertRecoveryAddress, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Account Recovery With Code Success", () => {
  ;[
    {
      recovery: react.recovery,
      base: react.base,
      app: "react" as "react",
      profile: "spa",
    },
    {
      recovery: express.recovery,
      base: express.base,
      app: "express" as "express",
      profile: "recovery",
    },
  ].forEach(({ recovery, profile, base, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      let identity

      beforeEach(() => {
        cy.deleteMail()
        cy.longRecoveryLifespan()
        cy.longLinkLifespan()
        cy.disableVerification()
        cy.enableRecovery()
        cy.useRecoveryStrategy("code")
        cy.notifyUnknownRecipients("recovery", false)

        identity = gen.identityWithWebsite()
        cy.registerApi(identity)
      })

      it("should contain the recovery address in the session", () => {
        cy.visit(recovery)
        cy.login({ ...identity, cookieUrl: base })
        cy.getSession().should(assertRecoveryAddress(identity))
      })

      it("should perform a recovery flow", () => {
        cy.visit(recovery)
        cy.get(appPrefix(app) + "input[name='email']").type(identity.email)
        cy.get("button[value='code']").click()
        cy.get('[data-testid="ui/message/1060003"]').should(
          "have.text",
          "An email containing a recovery code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
        )

        cy.recoveryEmailWithCode({ expect: { email: identity.email } })
        cy.get("button[value='code']").click()

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
        cy.expectSettingsSaved()
        cy.get('input[name="password"]').should("be.empty")

        cy.logout()
        cy.login({
          email: identity.email,
          password: newPassword,
          cookieUrl: base,
        })
      })

      it("should recover account with correct code after entering wrong code", () => {
        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.visit(recovery)
        cy.get(appPrefix(app) + "input[name='email']").type(identity.email)
        cy.get("button[value='code']").click()
        cy.get('[data-testid="ui/message/1060003"]').should(
          "have.text",
          "An email containing a recovery code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
        )
        cy.get("input[name='code']").type("12312312") // Invalid code
        cy.get("button[value='code']").click()
        cy.get('[data-testid="ui/message/4060006"]').should(
          "have.text",
          "The recovery code is invalid or has already been used. Please try again.",
        )
        cy.noSession()
        cy.recoveryEmailWithCode({ expect: { email: identity.email } })
        cy.get("button[value='code']").click()

        cy.get('[data-testid="ui/message/1060001"]', { timeout: 30000 }).should(
          "contain.text",
          "You successfully recovered your account. ",
        )
        cy.getSession()
        cy.location("pathname").should("eq", "/settings")
        cy.get('input[name="traits.email"]').should(
          "have.value",
          identity.email,
        )
      })

      it("should recover account after resending code", () => {
        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.visit(recovery)
        cy.get(appPrefix(app) + "input[name='email']").type(identity.email)
        cy.get("button[value='code']").click()
        cy.get('[data-testid="ui/message/1060003"]').should(
          "have.text",
          "An email containing a recovery code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
        )

        cy.recoveryEmailWithCode({
          expect: { email: identity.email, enterCode: false },
        })

        cy.get("button[name='email']").click() // resend code
        cy.noSession()

        cy.recoveryEmailWithCode({
          expect: { email: identity.email },
        })
        cy.get("button[value='code']").click()

        cy.get('[data-testid="ui/message/1060001"]', { timeout: 30000 }).should(
          "contain.text",
          "You successfully recovered your account. ",
        )
        cy.getSession()
        cy.location("pathname").should("eq", "/settings")
        cy.get('input[name="traits.email"]').should(
          "have.value",
          identity.email,
        )
      })
      it("should not notify an unknown recipient", () => {
        const recipient = gen.email()

        cy.visit(recovery)
        cy.get('input[name="email"]').type(recipient)
        cy.get(`[name="method"][value="code"]`).click()

        cy.getCourierMessages().then((messages) => {
          expect(messages.map((msg) => msg.recipient)).to.not.include(recipient)
        })
      })
    })
  })

  it("should recover, set password and be redirected", () => {
    const app = "express" as "express"
    cy.deleteMail()
    cy.useConfigProfile("recovery")
    cy.proxy(app)

    cy.deleteMail()
    cy.longRecoveryLifespan()
    cy.longCodeLifespan()
    cy.disableVerification()
    cy.enableRecovery()
    cy.useRecoveryStrategy("code")

    const identity = gen.identityWithWebsite()
    cy.registerApi(identity)
    cy.visit(express.recovery + "?return_to=https://www.ory.sh/")
    cy.get("input[name='email']").type(identity.email)
    cy.get("button[value='code']").click()
    cy.get('[data-testid="ui/message/1060003"]').should(
      "have.text",
      "An email containing a recovery code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
    )

    cy.recoveryEmailWithCode({ expect: { email: identity.email } })
    cy.get("button[value='code']").click()

    cy.getSession()
    cy.location("pathname").should("eq", "/settings")

    cy.get('input[name="password"]').clear().type(gen.password())
    cy.get('button[value="password"]').click()
    cy.url().should("eq", "https://www.ory.sh/")
  })
})

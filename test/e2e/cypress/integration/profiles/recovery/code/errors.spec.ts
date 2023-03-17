// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, email, extractRecoveryCode, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Account Recovery Errors", () => {
  ;[
    {
      recovery: react.recovery,
      app: "react" as "react",
      profile: "spa",
    },
    {
      recovery: express.recovery,
      app: "express" as "express",
      profile: "recovery",
    },
  ].forEach(({ recovery, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.deleteMail()
        cy.longRecoveryLifespan()
        cy.longCodeLifespan()
        cy.disableVerification()
        cy.enableRecovery()
        cy.useRecoveryStrategy("code")
        cy.notifyUnknownRecipients("recovery", false)
      })

      it("should invalidate flow if wrong code is submitted too often", () => {
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
        for (let i = 0; i < 5; i++) {
          cy.get("input[name='code']").type((i + "").repeat(8)) // Invalid code
          cy.get("button[value='code']").click()
          cy.get('[data-testid="ui/message/4060006"]').should(
            "have.text",
            "The recovery code is invalid or has already been used. Please try again.",
          )
          cy.noSession()
        }

        cy.get("input[name='code']").type("12312312") // Invalid code
        cy.get("button[value='code']").click()
        cy.get('[data-testid="ui/message/4000001"]').should(
          "have.text",
          "The request was submitted too often. Please request another code.",
        )
        cy.noSession()
        cy.get(appPrefix(app) + "input[name='email']").type(identity.email)
        cy.get("button[value='code']").click()
        cy.get('[data-testid="ui/message/1060003"]').should(
          "have.text",
          "An email containing a recovery code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
        )
        cy.recoveryEmailWithCode({
          expect: { email: identity.email, enterCode: false },
        })
      })

      it("shows code expired message if expired code is submitted", () => {
        cy.visit(recovery)

        cy.shortCodeLifespan()

        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.get(appPrefix(app) + "input[name='email']").type(identity.email)
        cy.get("button[value='code']").click()
        cy.recoveryEmailWithCode({ expect: { email: identity.email } })
        cy.get("button[value='code']").click()

        cy.get('[data-testid="ui/message/4060005"]').should(
          "contain.text",
          "The recovery flow expired",
        )

        cy.noSession()
      })

      it("should receive a stub email when recovering a non-existent account", () => {
        cy.notifyUnknownRecipients("recovery")
        cy.visit(recovery)

        const email = gen.email()
        cy.get(appPrefix(app) + 'input[name="email"]').type(email)
        cy.get('button[value="code"]').click()

        cy.location("pathname").should("eq", "/recovery")
        cy.get('[data-testid="ui/message/1060003"]').should(
          "have.text",
          "An email containing a recovery code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
        )
        cy.get('input[name="code"]').should("be.visible")

        cy.getMail().should((message) => {
          expect(message.subject).to.equal("Account access attempted")
          expect(message.fromAddress.trim()).to.equal("no-reply@ory.kratos.sh")
          expect(message.toAddresses).to.have.length(1)
          expect(message.toAddresses[0].trim()).to.equal(email)

          const code = extractRecoveryCode(message.body)
          expect(code).to.be.null
        })
      })

      it("should cause form errors", () => {
        cy.visit(recovery)
        cy.removeAttribute(["input[name='email']"], "required")
        cy.get('button[value="code"]').click()
        cy.get('[data-testid="ui/message/4000002"]').should(
          "contain.text",
          "Property email is missing.",
        )
        cy.get('[name="method"][value="code"]').should("exist")
      })

      it("is unable to recover the account if the code is incorrect", () => {
        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.visit(recovery)
        cy.get(appPrefix(app) + "input[name='email']").type(identity.email)
        cy.get("button[value='code']").click()
        cy.get('[data-testid="ui/message/1060003"]').should(
          "have.text",
          "An email containing a recovery code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
        )
        cy.get("input[name='code']").type("01234567") // Invalid code
        cy.get("button[value='code']").click()
        cy.get('[data-testid="ui/message/4060006"]').should(
          "have.text",
          "The recovery code is invalid or has already been used. Please try again.",
        )
        cy.noSession()
      })

      it("should cause non-repeating form errors after submitting empty form twice. see: #2512", () => {
        cy.visit(recovery)
        cy.location("pathname").should("eq", "/recovery")
        cy.removeAttribute(["input[name='email']"], "required")
        cy.get('button[value="code"]').click()
        cy.get('[data-testid="ui/message/4000002"]').should(
          "contain.text",
          "Property email is missing.",
        )
        cy.get("form")
          .find('[data-testid="ui/message/4000002"]')
          .should("have.length", 1)
        cy.get('[name="method"][value="code"]').should("exist")
      })

      it("remote recovery email template (recovery_code_valid)", () => {
        cy.remoteCourierRecoveryCodeTemplates()
        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.visit(recovery)
        cy.get(appPrefix(app) + "input[name='email']").type(identity.email)
        cy.get("button[value='code']").click()
        cy.get('[data-testid="ui/message/1060003"]').should(
          "have.text",
          "An email containing a recovery code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
        )

        cy.getMail().then((mail) => {
          expect(mail.body).to.include("recovery_code_valid REMOTE TEMPLATE")
        })
      })

      it("remote recovery email template (recovery_code_invalid)", () => {
        cy.notifyUnknownRecipients("recovery")
        cy.remoteCourierRecoveryCodeTemplates()
        cy.visit(recovery)
        cy.get(appPrefix(app) + "input[name='email']").type(email())
        cy.get("button[value='code']").click()
        cy.get('[data-testid="ui/message/1060003"]').should(
          "have.text",
          "An email containing a recovery code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
        )

        cy.getMail().then((mail) => {
          expect(mail.body).to.include("recovery_code_invalid REMOTE TEMPLATE")
        })
      })
    })
  })
})

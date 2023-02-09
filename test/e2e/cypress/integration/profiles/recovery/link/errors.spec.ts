// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, gen, parseHtml } from "../../../../helpers"
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
        cy.longLinkLifespan()
        cy.disableVerification()
        cy.enableRecovery()
        cy.useRecoveryStrategy("link")
        cy.disableRecoveryStrategy("code")
        cy.clearAllCookies()
        cy.notifyUnknownRecipients("verification", false)
      })

      it("responds with a HTML response on link click of an API flow if the link is expired", () => {
        cy.visit(recovery)

        cy.shortLinkLifespan()

        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.recoverApi({ email: identity.email })
        cy.recoverEmailButExpired({ expect: { email: identity.email } })

        cy.get('[data-testid="ui/message/4060005"]').should(
          "contain.text",
          "The recovery flow expired",
        )

        cy.noSession()
      })

      it("responds with a HTML response on link click of an API flow if the flow is expired", () => {
        cy.visit(recovery)

        cy.updateConfigFile((config) => {
          config.selfservice.flows.recovery.lifespan = "1s"
          return config
        })

        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.recoverApi({ email: identity.email })
        cy.wait(1000)

        cy.getMail().should((message) => {
          expect(message.subject).to.equal("Recover access to your account")
          expect(message.toAddresses[0].trim()).to.equal(identity.email)

          const link = parseHtml(message.body).querySelector("a")
          cy.longRecoveryLifespan()
          cy.visit(link.href)
        })

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
        cy.get('button[value="link"]').click()

        cy.location("pathname").should("eq", "/recovery")
        cy.get('[data-testid="ui/message/1060002"]').should(
          "have.text",
          "An email containing a recovery link has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
        )
        cy.get('input[name="email"]').should("have.value", email)

        cy.getMail().should((message) => {
          expect(message.subject).to.equal("Account access attempted")
          expect(message.fromAddress.trim()).to.equal("no-reply@ory.kratos.sh")
          expect(message.toAddresses).to.have.length(1)
          expect(message.toAddresses[0].trim()).to.equal(email)

          const link = parseHtml(message.body).querySelector("a")
          expect(link).to.be.null
        })
      })

      it("should cause form errors", () => {
        cy.visit(recovery)

        // we need to remove the required attribute of the element since the browser prevents us from submitting the form
        // this is to simulate the case where the form is submitted with an empty input field.
        cy.removeAttribute(["input[name='email']"], "required")

        cy.get('button[value="link"]').click()
        cy.get('[data-testid="ui/message/4000002"]').should(
          "contain.text",
          "Property email is missing.",
        )
        cy.get('[name="method"][value="link"]').should("exist")
      })

      it("should cause non-repeating form errors after submitting empty form twice. see: #2512", () => {
        cy.visit(recovery)
        // we need to remove the required attribute of the element since the browser prevents us from submitting the form
        // this is to simulate the case where the form is submitted with an empty input field.
        cy.removeAttribute(["input[name='email']"], "required")
        cy.get('button[value="link"]').click()
        cy.location("pathname").should("eq", "/recovery")

        // we need to remove the required attribute of the element since the browser prevents us from submitting the form
        // this is to simulate the case where the form is submitted with an empty input field.
        cy.removeAttribute(["input[name='email']"], "required")
        cy.get('button[value="link"]').click()
        cy.get('[data-testid="ui/message/4000002"]').should(
          "contain.text",
          "Property email is missing.",
        )
        cy.get("form")
          .find('[data-testid="ui/message/4000002"]')
          .should("have.length", 1)
        cy.get('[name="method"][value="link"]').should("exist")
      })

      it("is unable to recover the email address if the code is expired", () => {
        cy.shortLinkLifespan()
        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.recoverApi({ email: identity.email })
        cy.recoverEmailButExpired({ expect: { email: identity.email } })

        cy.get('[data-testid="ui/message/4060005"]').should(
          "contain.text",
          "The recovery flow expired",
        )

        cy.noSession()
      })

      it("is unable to recover the account if the code is incorrect", () => {
        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.recoverApi({ email: identity.email })

        cy.getMail().then((mail) => {
          console.log(mail)
          const link = parseHtml(mail.body).querySelector("a")
          cy.visit(link.href + "-not") // add random stuff to the confirm challenge
          cy.get('[data-testid="ui/message/4060004"]').should(
            "have.text",
            "The recovery token is invalid or has already been used. Please retry the flow.",
          )
          cy.noSession()
        })
      })

      it("is unable to recover the account using the token twice", () => {
        const identity = gen.identityWithWebsite()
        cy.registerApi(identity)
        cy.recoverApi({ email: identity.email })

        cy.getMail().then((mail) => {
          const link = parseHtml(mail.body).querySelector("a")

          // Workaround for cypress cy.visit limitation.
          cy.request(link.href).should((response) => {
            // add random stuff to the confirm challenge
            expect(response.status).to.eq(200)
          })

          cy.clearAllCookies()

          cy.visit(link.href)
          cy.get('[data-testid="ui/message/4060004"]').should(
            "have.text",
            "The recovery token is invalid or has already been used. Please retry the flow.",
          )
          cy.noSession()
        })
      })

      it("invalid remote recovery email template", () => {
        cy.notifyUnknownRecipients("recovery")
        cy.remoteCourierRecoveryTemplates()
        const identity = gen.identityWithWebsite()
        cy.recoverApi({ email: identity.email })

        cy.getMail().then((mail) => {
          expect(mail.body).to.include(
            "this is a remote invalid recovery template",
          )
        })
      })
    })
  })
})

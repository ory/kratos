// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import {
  assertVerifiableAddress,
  gen,
  parseHtml,
  verifyHrefPattern,
} from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"
import { Strategy } from "../../../../support"

context("Account Verification Error", () => {
  ;[
    {
      verification: react.verification,
      base: react.base,
      app: "react" as "react",
      profile: "verification",
    },
    {
      verification: express.verification,
      base: express.base,
      app: "express" as "express",
      profile: "verification",
    },
  ].forEach(({ profile, verification, app, base }) => {
    for (let s of ["code", "link"] as Strategy[]) {
      describe(`for strategy ${s}`, () => {
        describe(`for app ${app}`, () => {
          before(() => {
            cy.deleteMail()
            cy.useConfigProfile(profile)
            cy.proxy(app)
          })

          let identity
          beforeEach(() => {
            cy.clearAllCookies()
            cy.longVerificationLifespan()
            cy.longLifespan(s)
            cy.useVerificationStrategy(s)
            cy.resetCourierTemplates("verification")
            cy.notifyUnknownRecipients("verification", false)

            identity = gen.identity()
            cy.registerApi(identity)
            cy.deleteMail({ atLeast: 1 }) // clean up registration email
            cy.login(identity)
            cy.visit(verification)
          })

          it("responds with a HTML response on link click of an API flow if the flow is expired", () => {
            cy.updateConfigFile((config) => {
              config.selfservice.flows.verification.lifespan = "1s"
              return config
            })

            cy.verificationApi({
              email: identity.email,
              strategy: s,
            })

            cy.wait(1000)
            cy.shortVerificationLifespan()

            cy.getMail().then((message) => {
              expect(message.subject).to.equal(
                "Please verify your email address",
              )
              expect(message.toAddresses[0].trim()).to.equal(identity.email)

              const link = parseHtml(message.body).querySelector("a")

              cy.longVerificationLifespan()
              cy.visit(link.href)
              cy.get('[data-testid="ui/message/4070005"]').should(
                "contain.text",
                "verification flow expired",
              )

              cy.getSession().should((session) => {
                assertVerifiableAddress({
                  isVerified: false,
                  email: identity.email,
                })(session)
              })
            })
          })

          it("responds with a HTML response on link click of an API flow if the code is expired", () => {
            cy.shortLifespan(s)

            // Init expired flow
            cy.verificationApi({
              email: identity.email,
              strategy: s,
            })

            cy.verifyEmailButExpired({
              expect: { email: identity.email },
              strategy: s,
            })
          })

          it("is unable to verify the email address if the code is expired", () => {
            cy.shortLifespan(s)

            cy.visit(verification)
            cy.get('input[name="email"]').type(identity.email)
            cy.get(`button[value="${s}"]`).click()

            cy.contains("An email containing a verification")
            cy.get(`[name="method"][value="${s}"]`).should("exist")
            cy.verifyEmailButExpired({
              expect: { email: identity.email },
              strategy: s,
            })
          })

          it("is unable to verify the email address if the code is incorrect", () => {
            cy.get('input[name="email"]').type(identity.email)
            cy.get(`button[value="${s}"]`).click()

            cy.contains("An email containing a verification")

            cy.getMail().then((mail) => {
              const link = parseHtml(mail.body).querySelector("a")

              expect(verifyHrefPattern.test(link.href)).to.be.true

              cy.visit(link.href + "-not") // add random stuff to the confirm challenge
              cy.getSession().then(
                assertVerifiableAddress({
                  isVerified: false,
                  email: identity.email,
                }),
              )
            })
          })

          it("unable to verify non-existent account", () => {
            cy.notifyUnknownRecipients("verification")
            const email = gen.identity().email
            cy.get('input[name="email"]').type(email)
            cy.get(`button[value="${s}"]`).click()
            cy.getMail().then((mail) => {
              expect(mail.toAddresses).includes(email)
              expect(mail.subject).eq(
                "Someone tried to verify this email address",
              )
            })
          })

          if (s === "code") {
            it("is unable to brute force the code", () => {
              cy.visit(verification)
              cy.get('input[name="email"]').type(identity.email)
              cy.get(`button[value="${s}"]`).click()

              cy.contains("An email containing a verification")

              for (let i = 0; i < 5; i++) {
                cy.get("input[name='code']").type((i + "").repeat(8)) // Invalid code
                cy.get("button[value='code']").click()
                cy.get('[data-testid="ui/message/4070006"]').should(
                  "have.text",
                  "The verification code is invalid or has already been used. Please try again.",
                )
              }

              cy.get("input[name='code']").type("12312312") // Invalid code
              cy.get("button[value='code']").click()
              cy.get('[data-testid="ui/message/4000001"]').should(
                "have.text",
                "The request was submitted too often. Please request another code.",
              )
              cy.getSession().then(
                assertVerifiableAddress({
                  isVerified: false,
                  email: identity.email,
                }),
              )
            })
          }
        })
      })
    }
  })
})

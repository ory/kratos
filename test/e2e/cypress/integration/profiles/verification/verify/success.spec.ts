// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { APP_URL, assertVerifiableAddress, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"
import { Strategy } from "../../../../support"

context("Account Verification Settings Success", () => {
  ;[
    {
      verification: react.verification,
      app: "react" as "react",
      profile: "verification",
    },
    {
      verification: express.verification,
      app: "express" as "express",
      profile: "verification",
    },
  ].forEach(({ profile, verification, app }) => {
    describe(`for app ${app}`, () => {
      for (let s of ["code", "link"] as Strategy[]) {
        describe(`for strategy ${s}`, () => {
          before(() => {
            cy.deleteMail()
            cy.useConfigProfile(profile)
            cy.proxy(app)
          })

          let identity

          beforeEach(() => {
            cy.useVerificationStrategy(s)
            cy.notifyUnknownRecipients("verification", false)
            identity = gen.identity()
            cy.register(identity)
            cy.deleteMail({ atLeast: 1 }) // clean up registration email

            cy.login(identity)
            cy.visit(verification)
          })

          it("should request verification and receive an email and verify it", () => {
            cy.get('input[name="email"]').type(identity.email)
            cy.get(`button[value="${s}"]`).click()

            cy.contains("An email containing a verification")

            cy.get(`[name="method"][value="${s}"]`).should("exist")

            cy.verifyEmail({ expect: { email: identity.email }, strategy: s })
          })

          it("should request verification for an email that does not exist yet", () => {
            cy.notifyUnknownRecipients("verification")
            const email = `not-${identity.email}`
            cy.get('input[name="email"]').type(email)
            cy.get(`button[value="${s}"]`).click()

            cy.contains("An email containing a verification")

            cy.getMail().should((message) => {
              expect(message.subject.trim()).to.equal(
                "Someone tried to verify this email address",
              )
              expect(message.fromAddress.trim()).to.equal(
                "no-reply@ory.kratos.sh",
              )
              expect(message.toAddresses).to.have.length(1)
              expect(message.toAddresses[0].trim()).to.equal(email)
            })

            cy.getSession().then(
              assertVerifiableAddress({
                isVerified: false,
                email: identity.email,
              }),
            )
          })

          it("should not verify email when clicking on link received on different address", () => {
            cy.get('input[name="email"]').type(identity.email)
            cy.get(`button[value="${s}"]`).click()

            cy.verifyEmail({ expect: { email: identity.email }, strategy: s })

            // identity is verified
            cy.logout()

            // registered with other email address
            const identity2 = gen.identity()
            cy.register(identity2)
            cy.deleteMail({ atLeast: 1 }) // clean up registration email

            cy.login(identity2)

            cy.visit(APP_URL + "/verification")

            // request verification link for identity
            cy.get('input[name="email"]').type(identity.email)
            cy.get(`button[value="${s}"]`).click()

            cy.performEmailVerification({
              expect: { email: identity.email },
              strategy: s,
            })

            // expect current session to still not have a verified email address
            cy.getSession().should(
              assertVerifiableAddress({
                email: identity2.email,
                isVerified: false,
              }),
            )
          })

          it("should redirect to return_to after completing verification", () => {
            cy.clearAllCookies()
            // registered with other email address
            const identity2 = gen.identity()
            cy.register(identity2)
            cy.deleteMail({ atLeast: 1 }) // clean up registration email

            cy.login(identity2)

            cy.visit(APP_URL + "/self-service/verification/browser", {
              qs: { return_to: "http://localhost:4455/verification_callback" },
            })
            // request verification link for identity
            cy.get('input[name="email"]').type(identity2.email)
            cy.get(`[name="method"][value="${s}"]`).click()
            cy.verifyEmail({
              expect: {
                email: identity2.email,
                redirectTo: "http://localhost:4455/verification_callback",
              },
              strategy: s,
            })
          })

          it("should not notify an unknown recipient", () => {
            const recipient = gen.email()

            cy.visit(APP_URL + "/self-service/verification/browser")
            cy.get('input[name="email"]').type(recipient)
            cy.get(`[name="method"][value="${s}"]`).click()

            cy.getCourierMessages().then((messages) => {
              expect(messages.map((msg) => msg.recipient)).to.not.include(
                recipient,
              )
            })
          })
        })
      }
    })
  })
})

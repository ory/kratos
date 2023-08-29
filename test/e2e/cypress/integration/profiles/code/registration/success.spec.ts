// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Registration success with code method", () => {
  ;[
    {
      route: express.registration,
      login: express.login,
      recovery: express.recovery,
      app: "express" as "express",
      profile: "code",
    },
    {
      route: react.registration,
      login: react.login,
      recovery: react.recovery,
      app: "react" as "react",
      profile: "code",
    },
  ].forEach(({ route, login, recovery, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.deleteMail()
        cy.clearAllCookies()
        cy.visit(route)
      })

      it("should be able to resend the registration code", async () => {
        const email = gen.email()

        cy.get(`input[name='traits.email']`).type(email)

        cy.submitCodeForm()
        cy.get('[data-testid="ui/message/1040005"]').should(
          "contain",
          "An email containing a code has been sent to the email address you provided",
        )

        cy.getRegistrationCodeFromEmail(email).should((code) =>
          cy.wrap(code).as("code1"),
        )

        cy.get(`input[name='traits.email']`).should("have.value", email)
        cy.get(`input[name='method'][value='code'][type='hidden']`).should(
          "exist",
        )
        cy.get(`button[name='resend'][value='code']`).click()

        cy.getRegistrationCodeFromEmail(email).should((code) => {
          cy.wrap(code).as("code2")
        })

        cy.get("@code1").then((code1) => {
          // previous code should not work
          cy.get('input[name="code"]').clear().type(code1.toString())

          cy.submitCodeForm()
          cy.get('[data-testid="ui/message/4040003"]').should(
            "contain.text",
            "The registration code is invalid or has already been used. Please try again.",
          )
        })

        cy.get("@code2").then((code2) => {
          cy.get('input[name="code"]').clear().type(code2.toString())
          cy.submitCodeForm()
        })

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.verifiable_addresses).to.have.length(1)
          expect(identity.verifiable_addresses[0].status).to.equal("completed")
          expect(identity.traits.email).to.equal(email)
        })
      })

      it("should sign up and be logged in with session hook", () => {
        const email = gen.email()

        cy.get(` input[name='traits.email']`).type(email)

        cy.submitCodeForm()
        cy.get('[data-testid="ui/message/1040005"]').should(
          "contain",
          "An email containing a code has been sent to the email address you provided",
        )

        cy.getRegistrationCodeFromEmail(email).should((code) => {
          cy.get(`input[name=code]`).type(code)
          cy.get("button[name=method][value=code]").click()
        })

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.verifiable_addresses).to.have.length(1)
          expect(identity.verifiable_addresses[0].status).to.equal("completed")
          expect(identity.traits.email).to.equal(email)
        })
      })

      it("should be able to sign up without session hook", () => {
        cy.setPostCodeRegistrationHooks([])
        const email = gen.email()

        cy.get(`input[name='traits.email']`).type(email)

        cy.submitCodeForm()
        cy.get('[data-testid="ui/message/1040005"]').should(
          "contain",
          "An email containing a code has been sent to the email address you provided",
        )

        cy.getRegistrationCodeFromEmail(email).should((code) => {
          cy.get(`input[name=code]`).type(code)
          cy.get("button[name=method][value=code]").click()
        })

        cy.visit(login)
        cy.get(`input[name=identifier]`).type(email)
        cy.get("button[name=method][value=code]").click()

        cy.getLoginCodeFromEmail(email).then((code) => {
          cy.get(`input[name=code]`).type(code)
          cy.get("button[name=method][value=code]").click()
        })

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.verifiable_addresses).to.have.length(1)
          expect(identity.verifiable_addresses[0].status).to.equal("completed")
          expect(identity.traits.email).to.equal(email)
        })
      })

      it("should be able to recover account when registered with code", () => {
        const email = gen.email()
        cy.registerWithCode({ email })

        cy.clearAllCookies()
        cy.visit(recovery)

        cy.get('input[name="email"]').type(email)
        cy.get('button[name="method"][value="code"]').click()

        cy.recoveryEmailWithCode({ expect: { email } })
        cy.get('button[value="code"]').click()

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.traits.email).to.equal(email)
        })
      })

      // Try keep this test as the last one, as it updates the identity schema.
      it("should be able to use multiple identifiers to signup with and sign in to", () => {
        cy.setPostCodeRegistrationHooks([
          {
            hook: "session",
          },
        ])

        // Setup complex schema
        cy.setIdentitySchema(
          "file://test/e2e/profiles/code/identity.complex.traits.schema.json",
        )

        cy.visit(route)

        cy.get(`input[name='traits.username']`).type(Math.random().toString(36))

        const email = gen.email()
        cy.get(`input[name='traits.email']`).type(email)

        const email2 = gen.email()
        cy.get(`input[name='traits.email2']`).type(email2)

        cy.submitCodeForm()
        cy.get('[data-testid="ui/message/1040005"]').should(
          "contain",
          "An email containing a code has been sent to the email address you provided",
        )

        // intentionally use email 1 to sign up for the account
        cy.getRegistrationCodeFromEmail(email, { expectedCount: 1 }).should(
          (code) => {
            cy.get(`input[name=code]`).type(code)
            cy.get("button[name=method][value=code]").click()
          },
        )

        cy.logout()

        // There are verification emails from the registration process in the inbox that we need to deleted
        // for the assertions below to pass.
        cy.deleteMail({ atLeast: 1 })

        // Attempt to sign in with email 2 (should fail)
        cy.visit(login)
        cy.get(`input[name=identifier]`).type(email2)

        cy.get("button[name=method][value=code]").click()

        cy.getLoginCodeFromEmail(email2, {
          expectedCount: 1,
        }).should((code) => {
          cy.get(`input[name=code]`).type(code)
          cy.get("button[name=method][value=code]").click()
        })

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.verifiable_addresses).to.have.length(2)
          expect(
            identity.verifiable_addresses.filter((v) => v.value === email)[0]
              .status,
          ).to.equal("completed")
          expect(
            identity.verifiable_addresses.filter((v) => v.value === email2)[0]
              .status,
          ).to.equal("completed")
          expect(identity.traits.email).to.equal(email)
        })
      })
    })
  })
})

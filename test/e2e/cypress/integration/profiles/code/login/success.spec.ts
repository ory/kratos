// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Login success with code method", () => {
  ;[
    {
      route: express.login,
      app: "express" as "express",
      profile: "code",
    },
    {
      route: react.login,
      app: "react" as "react",
      profile: "code",
    },
  ].forEach(({ route, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
        cy.setPostCodeRegistrationHooks([])
        cy.setupHooks("login", "after", "code", [])
      })

      beforeEach(() => {
        const email = gen.email()
        cy.wrap(email).as("email")
        cy.registerWithCode({ email })

        cy.deleteMail()
        cy.clearAllCookies()
        cy.visit(route)
      })

      it("should be able to sign in with code", () => {
        cy.get("@email").then((email) => {
          cy.get('input[name="identifier"]').clear().type(email.toString())
          cy.submitCodeForm()

          cy.getLoginCodeFromEmail(email.toString()).should((code) => {
            cy.get('input[name="code"]').type(code)

            cy.get("button[name=method][value=code]").click()
          })

          cy.location("pathname").should("not.contain", "login")

          cy.getSession().should((session) => {
            const { identity } = session
            expect(identity.id).to.not.be.empty
            expect(identity.verifiable_addresses).to.have.length(1)
            expect(identity.verifiable_addresses[0].status).to.equal(
              "completed",
            )
            expect(identity.traits.email).to.equal(email)
          })
        })
      })

      it("should be able to resend login code", () => {
        cy.get("@email").then((email) => {
          cy.get('input[name="identifier"]').clear().type(email.toString())
          cy.submitCodeForm()

          cy.getLoginCodeFromEmail(email.toString()).should((code) => {
            cy.wrap(code).as("code1")
          })

          cy.get("button[name=resend]").click()

          cy.getLoginCodeFromEmail(email.toString()).should((code) => {
            cy.wrap(code).as("code2")
          })

          cy.get("@code1").then((code1) => {
            cy.get("@code2").then((code2) => {
              expect(code1).to.not.equal(code2)
            })
          })

          // attempt to submit code 1
          cy.get("@code1").then((code1) => {
            cy.get('input[name="code"]').clear().type(code1.toString())
          })

          cy.get("button[name=method][value=code]").click()

          cy.get("[data-testid='ui/message/4010008']").contains(
            "The login code is invalid or has already been used",
          )

          // attempt to submit code 2
          cy.get("@code2").then((code2) => {
            cy.get('input[name="code"]').clear().type(code2.toString())
          })

          cy.get('button[name="method"][value="code"]').click()

          if (app === "express") {
            cy.get('a[href*="sessions"').click()
          }
          cy.getSession().should((session) => {
            const { identity } = session
            expect(identity.id).to.not.be.empty
            expect(identity.verifiable_addresses).to.have.length(1)
            expect(identity.verifiable_addresses[0].status).to.equal(
              "completed",
            )
            expect(identity.traits.email).to.equal(email)
          })
        })
      })

      it("should be able to login to un-verfied email", () => {
        const email = gen.email()
        const email2 = gen.email()

        // Setup complex schema
        cy.setIdentitySchema(
          "file://test/e2e/profiles/code/identity.complex.traits.schema.json",
        )

        cy.registerWithCode({
          email: email,
          traits: {
            "traits.username": Math.random().toString(36),
            "traits.email2": email2,
          },
        })

        // There are verification emails from the registration process in the inbox that we need to deleted
        // for the assertions below to pass.
        cy.deleteMail({ atLeast: 1 })

        cy.visit(route)

        cy.get('input[name="identifier"]').clear().type(email2)
        cy.submitCodeForm()

        cy.getLoginCodeFromEmail(email2).should((code) => {
          cy.get('input[name="code"]').type(code)
          cy.get("button[name=method][value=code]").click()
        })

        cy.getSession({ expectAal: "aal1", expectMethods: ["code"] }).then(
          (session) => {
            expect(session.identity.verifiable_addresses).to.have.length(2)
            expect(session.identity.verifiable_addresses[0].status).to.equal(
              "completed",
            )
            expect(session.identity.verifiable_addresses[1].status).to.equal(
              "completed",
            )
          },
        )
      })
    })
  })
})

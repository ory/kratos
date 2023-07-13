// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { should } from "chai"
import { appPrefix, APP_URL, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Registration success with code method", () => {
  ;[
    {
      route: express.registration,
      login: express.login,
      app: "express" as "express",
      profile: "code",
    },
    // {
    //   route: react.registration,
    //   app: "react" as "react",
    //   profile: "code",
    // },
  ].forEach(({ route, login, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        cy.proxy(app)
        cy.setIdentitySchema(
          "file://test/e2e/profiles/code/identity.traits.schema.json",
        )
        cy.setPostCodeRegistrationHooks([])
        cy.setupHooks("login", "after", "code", [])
      })

      beforeEach(() => {
        cy.deleteMail()
        cy.clearAllCookies()
        cy.visit(route)
      })

      it("should be able to sign up without session hook", () => {
        const email = gen.email()

        cy.get(
          "form[data-testid='registration-flow-code'] input[name='traits.email']",
        ).type(email)

        cy.submitCodeForm()

        cy.url().should("contain", "registration")
        cy.getRegistrationCodeFromEmail(email).then((code) => {
          cy.get(
            "form[data-testid='registration-flow-code'] input[name=code]",
          ).type(code)
          cy.get("button[name=method][value=code]").click()
        })

        cy.deleteMail({ atLeast: 1 })

        cy.visit(login)
        cy.get(
          "form[data-testid='login-flow-code'] input[name=identifier]",
        ).type(email)
        cy.get("button[name=method][value=code]").click()

        cy.getLoginCodeFromEmail(email).then((code) => {
          cy.get("form[data-testid='login-flow-code'] input[name=code]").type(
            code,
          )
          cy.get("button[name=method][value=code]").click()
        })

        cy.deleteMail({ atLeast: 1 })

        if (app === "express") {
          cy.get('a[href*="sessions"').click()
        }
        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.verifiable_addresses).to.have.length(1)
          expect(identity.verifiable_addresses[0].status).to.equal("completed")
          expect(identity.traits.email).to.equal(email)
        })
      })

      it("should be able to resend the registration code", async () => {
        cy.setPostCodeRegistrationHooks([
          {
            hook: "session",
          },
        ])
        const email = gen.email()

        cy.get(
          "form[data-testid='registration-flow-code'] input[name='traits.email']",
        ).type(email)

        cy.submitCodeForm()

        cy.url().should("contain", "registration")

        cy.getRegistrationCodeFromEmail(email).then((code) =>
          cy.wrap(code).as("code1"),
        )

        cy.get(
          "form[data-testid='registration-flow-code'] input[name='traits.email']",
        ).should("have.value", email)
        cy.get(
          "form[data-testid='registration-flow-code'] input[name='method'][value='code'][type='hidden']",
        ).should("exist")
        cy.get(
          "form[data-testid='registration-flow-code'] button[name='resend'][value='code']",
        ).click()

        cy.getRegistrationCodeFromEmail(email).then((code) => {
          cy.wrap(code).as("code2")
        })

        cy.get("@code1").then((code1) => {
          // previous code should not work
          cy.get(
            'form[data-testid="registration-flow-code"] input[name="code"]',
          )
            .clear()
            .type(code1.toString())
          cy.submitCodeForm()

          cy.get('[data-testid="ui/message/4040003"]').should(
            "contain.text",
            "The registration code is invalid or has already been used. Please try again.",
          )
        })

        cy.get("@code2").then((code2) => {
          cy.get(
            'form[data-testid="registration-flow-code"] input[name="code"]',
          )
            .clear()
            .type(code2.toString())
          cy.submitCodeForm()
        })

        if (app === "express") {
          cy.get('a[href*="sessions"').click()
        }
        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.verifiable_addresses).to.have.length(1)
          expect(identity.verifiable_addresses[0].status).to.equal("completed")
          expect(identity.traits.email).to.equal(email)
        })
      })

      it("should sign up and be logged in with session hook", () => {
        cy.setPostCodeRegistrationHooks([
          {
            hook: "session",
          },
        ])

        const email = gen.email()

        cy.get(
          "form[data-testid='registration-flow-code'] input[name='traits.email']",
        ).type(email)

        cy.submitCodeForm()

        cy.url().should("contain", "registration")
        cy.getRegistrationCodeFromEmail(email).then((code) => {
          cy.get(
            "form[data-testid='registration-flow-code'] input[name=code]",
          ).type(code)
          cy.get("button[name=method][value=code]").click()
        })

        cy.deleteMail({ atLeast: 1 })

        if (app === "express") {
          cy.get('a[href*="sessions"').click()
        }
        cy.get("pre").should("contain.text", email)

        cy.getSession().should((session) => {
          const { identity } = session
          expect(identity.id).to.not.be.empty
          expect(identity.verifiable_addresses).to.have.length(1)
          expect(identity.verifiable_addresses[0].status).to.equal("completed")
          expect(identity.traits.email).to.equal(email)
        })
      })

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

        cy.get(
          "form[data-testid='registration-flow-code'] input[name='traits.username']",
        ).type(Math.random().toString(36))

        const email = gen.email()

        cy.get(
          "form[data-testid='registration-flow-code'] input[name='traits.email']",
        ).type(email)

        const email2 = gen.email()

        cy.get(
          "form[data-testid='registration-flow-code'] input[name='traits.email2']",
        ).type(email2)

        cy.submitCodeForm()

        // intentionally use email 1 to verify the account
        cy.url().should("contain", "registration")
        cy.getRegistrationCodeFromEmail(email, { expectedCount: 2 }).then(
          (code) => {
            cy.get(
              "form[data-testid='registration-flow-code'] input[name=code]",
            ).type(code)
            cy.get("button[name=method][value=code]").click()
          },
        )

        cy.deleteMail({ atLeast: 2 })

        cy.logout()

        // Attempt to sign in with email 2 (should fail)
        cy.visit(login)
        cy.get(
          "form[data-testid='login-flow-code'] input[name=identifier]",
        ).type(email2)

        cy.get("button[name=method][value=code]").click()

        cy.getLoginCodeFromEmail(email2).then((code) => {
          cy.get("form[data-testid='login-flow-code'] input[name=code]").type(
            code,
          )
          cy.get("button[name=method][value=code]").click()
        })
        if (app === "express") {
          cy.get('a[href*="sessions"').click()
        }

        cy.getSession().should((session) => {
          console.dir({ session })
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

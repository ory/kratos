// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { Session } from "@ory/kratos-client"
import { gen, MOBILE_URL } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

const Selectors = {
  mobile: {
    identifier: "[data-testid='field/identifier']",
    recoveryEmail: "[data-testid='field/email']",
    email: "[data-testid='traits.email']",
    email2: "[data-testid='traits.email2']",
    website: "[data-testid='traits.website']",
    username: "[data-testid='traits.username']",
    code: "[data-testid='field/code'] input",
    recoveryCode: "[data-testid='code']",
    submitCode: "[data-testid='field/method/code']",
    resendCode: "[data-testid='field/resend/code']",
    credentialSelection: "[data-testid='field/screen/credential-selection']",
    submitRecovery: "[data-testid='field/method/code']",
    codeHiddenMethod: "[data-testid='field/method/code']",
  },
  express: {
    identifier: "[data-testid='login-flow-code'] input[name='identifier']",
    recoveryEmail: "input[name=email]",
    email: "[data-testid='node/input/traits.email'] input[name='traits.email']",
    email2:
      "[data-testid='node/input/traits.email2'] input[name='traits.email2']",
    website:
      "[data-testid='node/input/traits.website'] [name='traits.website']",
    username:
      "[data-testid='node/input/traits.username'] input[name='traits.username']",
    code: "input[name='code']",
    recoveryCode: "input[name=code]",
    submitRecovery: "button[name=method][value=code]",
    submitCode: "button[name='method'][value='code']",
    resendCode: "button[name='resend'][value='code']",
    codeHiddenMethod: "input[name='method'][value='code'][type='hidden']",
    credentialSelection: "[name='screen'][value='credential-selection']",
  },
  react: {
    identifier: "input[name='identifier']",
    recoveryEmail: "input[name=email]",
    email: "input[name='traits.email']",
    email2: "input[name='traits.email2']",
    website: "[name='traits.website']",
    username: "input[name='traits.username']",
    code: "input[name='code']",
    recoveryCode: "input[name=code]",
    submitRecovery: "button[name=method][value=code]",
    submitCode: "button[name='method'][value='code']",
    resendCode: "button[name='resend'][value='code']",
    codeHiddenMethod: "input[name='method'][value='code'][type='hidden']",
    credentialSelection: "[name='screen'][value='credential-selection']",
  },
}

context("Registration success with code method", () => {
  ;[
    {
      route: express.registration,
      login: express.login,
      recovery: express.recovery,
      app: "express" as "express",
      profile: "two-steps",
    },
    {
      route: react.registration,
      login: react.login,
      recovery: react.recovery,
      app: "react" as "react",
      profile: "two-steps",
    },
    {
      route: MOBILE_URL + "/Registration",
      login: MOBILE_URL + "/Login",
      recovery: MOBILE_URL + "/Recovery",
      app: "mobile" as "mobile",
      profile: "two-steps",
    },
  ].forEach(({ route, login, recovery, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        if (app !== "mobile") {
          cy.proxy(app)
        }
      })

      beforeEach(() => {
        cy.deleteMail({ atLeast: 0 })
        cy.clearAllCookies()
        cy.visit(route)
      })

      it("should be able to resend the registration code", () => {
        const email = gen.email()
        const website = "https://www.example.org/"

        cy.get(Selectors[app]["email"]).type(email)
        cy.get(Selectors[app]["website"]).type(website)

        cy.submitProfileForm(app)
        cy.submitCodeForm(app)
        cy.get('[data-testid="ui/message/1040005"]').should("be.visible")

        cy.getRegistrationCodeFromEmail(email).then((code) =>
          cy.wrap(code).as("code1"),
        )

        // cy.get(Selectors[app]["email"]).should("have.value", email)
        cy.get(Selectors[app]["codeHiddenMethod"]).should("exist")
        cy.get(Selectors[app]["resendCode"]).click()

        cy.getRegistrationCodeFromEmail(email).then((code) => {
          cy.wrap(code).as("code2")
        })

        cy.get("@code1").then((code1) => {
          // previous code should not work
          cy.get(Selectors[app]["code"]).clear()
          cy.get(Selectors[app]["code"]).type(code1.toString())

          cy.submitCodeForm(app)
          cy.get('[data-testid="ui/message/4040003"]').should(
            "contain.text",
            "The registration code is invalid or has already been used. Please try again.",
          )
        })

        // Navigate back and forth again
        cy.get(Selectors[app]["credentialSelection"]).click()
        cy.submitCodeForm(app)

        cy.getRegistrationCodeFromEmail(email).then((code) => {
          cy.wrap(code).as("code2")
        })

        cy.get("@code2").then((code2) => {
          cy.get(Selectors[app]["code"]).clear()
          cy.get(Selectors[app]["code"]).type(code2.toString())
          cy.submitCodeForm(app)
        })

        if (app === "express") {
          cy.url().should("match", /\/welcome/)
        } else {
          cy.get('[data-testid="session-content"]').should("contain", email)
        }

        if (app === "mobile") {
          cy.get('[data-testid="session-token"]').should("not.be.empty")
          cy.get('[data-testid="session-token"]').then((token) => {
            cy.getSession({
              expectAal: "aal1",
              expectMethods: ["code"],
              token: token.text(),
            }).then((session) => {
              cy.wrap(session).as("session")
            })
          })
        } else {
          cy.getSession({ expectAal: "aal1", expectMethods: ["code"] }).then(
            (session) => {
              cy.wrap(session).as("session")
            },
          )
        }

        cy.get<Session>("@session").then(({ identity }) => {
          expect(identity.id).to.not.be.empty
          expect(identity.traits.email).to.equal(email)
        })
      })

      it("should sign up and be logged in with session hook", () => {
        const email = gen.email()
        const website = "https://www.example.org/"

        cy.get(Selectors[app]["email"]).type(email)
        cy.get(Selectors[app]["website"]).type(website)
        cy.submitProfileForm(app)

        cy.submitCodeForm(app)
        cy.get('[data-testid="ui/message/1040005"]').should("be.visible")

        cy.getRegistrationCodeFromEmail(email).should((code) => {
          cy.get(Selectors[app]["code"]).type(code)
          cy.get(Selectors[app]["submitCode"]).click()
        })

        if (app === "mobile") {
          cy.get('[data-testid="session-token"]').then((token) => {
            cy.getSession({
              expectAal: "aal1",
              expectMethods: ["code"],
              token: token.text(),
            }).then((session) => {
              cy.wrap(session).as("session")
            })
          })

          cy.get('[data-testid="session-content"]').should("contain", email)
          cy.get('[data-testid="session-token"]').should("not.be.empty")
        } else {
          cy.getSession({ expectAal: "aal1", expectMethods: ["code"] }).then(
            (session) => {
              cy.wrap(session).as("session")
            },
          )
        }

        cy.get<Session>("@session").then(({ identity }) => {
          expect(identity.id).to.not.be.empty
          expect(identity.traits.email).to.equal(email)
        })
      })

      it("should be able to sign up without session hook", () => {
        cy.setPostCodeRegistrationHooks([])
        const email = gen.email()
        const website = "https://www.example.org/"

        cy.get(Selectors[app]["email"]).type(email)
        cy.get(Selectors[app]["website"]).type(website)
        cy.submitProfileForm(app)

        cy.submitCodeForm(app)
        cy.get('[data-testid="ui/message/1040005"]').should("be.visible")

        cy.getRegistrationCodeFromEmail(email).should((code) => {
          cy.get(Selectors[app]["code"]).type(code)
          cy.get(Selectors[app]["submitCode"]).click()
        })

        cy.visit(login)
        cy.get(Selectors[app]["identifier"]).type(email)
        cy.get(Selectors[app]["submitCode"]).click()

        cy.getLoginCodeFromEmail(email).then((code) => {
          cy.get(Selectors[app]["code"]).type(code)
          cy.get(Selectors[app]["submitCode"]).click()
        })

        if (app === "mobile") {
          cy.get('[data-testid="session-token"]').then((token) => {
            cy.getSession({
              expectAal: "aal1",
              expectMethods: ["code"],
              token: token.text(),
            }).then((session) => {
              cy.wrap(session).as("session")
            })
          })

          cy.get('[data-testid="session-content"]').should("contain", email)
          cy.get('[data-testid="session-token"]').should("not.be.empty")
        } else {
          cy.getSession({ expectAal: "aal1", expectMethods: ["code"] }).then(
            (session) => {
              cy.wrap(session).as("session")
            },
          )
        }

        cy.get<Session>("@session").then(({ identity }) => {
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

        cy.get(Selectors[app]["username"]).type(Math.random().toString(36))

        const email = gen.email()
        cy.get(Selectors[app]["email"]).type(email)

        const email2 = gen.email()
        cy.get(Selectors[app]["email2"]).type(email2)

        cy.submitProfileForm(app)
        cy.submitCodeForm(app)
        cy.get('[data-testid="ui/message/1040005"]').should("be.visible")

        // intentionally use email 1 to sign up for the account
        cy.getRegistrationCodeFromEmail(email, { expectedCount: 1 }).should(
          (code) => {
            cy.get(Selectors[app]["code"]).type(code)
            cy.get(Selectors[app]["submitCode"]).click()
          },
        )

        if (app === "mobile") {
          cy.visit(MOBILE_URL + "/Home")
          cy.get('*[data-testid="logout"]').click()
        } else {
          cy.logout()
        }

        // There are verification emails from the registration process in the inbox that we need to deleted
        // for the assertions below to pass.
        cy.deleteMail({ atLeast: 1 })

        // Attempt to sign in with email 2 (should fail)
        cy.visit(login)
        cy.get(Selectors[app]["identifier"]).type(email2)

        cy.get(Selectors[app]["submitCode"]).click()

        cy.getLoginCodeFromEmail(email2, {
          expectedCount: 1,
        }).should((code) => {
          cy.get(Selectors[app]["code"]).type(code)
          cy.get(Selectors[app]["submitCode"]).click()
        })

        if (app === "mobile") {
          cy.get('[data-testid="session-token"]').then((token) => {
            cy.getSession({
              expectAal: "aal1",
              expectMethods: ["code"],
              token: token.text(),
            }).then((session) => {
              cy.wrap(session).as("session")
            })
          })

          cy.get('[data-testid="session-content"]').should("contain", email)
          cy.get('[data-testid="session-token"]').should("not.be.empty")
        } else {
          cy.getSession({ expectAal: "aal1", expectMethods: ["code"] }).then(
            (session) => {
              cy.wrap(session).as("session")
            },
          )
        }

        cy.get<Session>("@session").then(({ identity }) => {
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

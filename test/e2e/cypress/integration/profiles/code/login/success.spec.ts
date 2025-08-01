// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { Session } from "@ory/kratos-client"
import { gen, MOBILE_URL } from "../../../../helpers"
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
    {
      route: MOBILE_URL + "/Login",
      app: "mobile" as "mobile",
      profile: "code",
    },
  ].forEach(({ route, profile, app }) => {
    describe(`for app ${app}`, () => {
      const Selectors = {
        mobile: {
          identity: '[data-testid="identifier"]',
          code: '[data-testid="code"]',
          submit: '[data-testid="field/method/code"]',
          resend: '[data-testid="field/resend/code"]',
        },
        express: {
          identity: '[data-testid="login-flow-code"] input[name="identifier"]',
          code: 'input[name="code"]',
          submit: 'button[name="method"][value="code"]',
          resend: 'button[name="resend"]',
        },
        react: {
          identity: 'input[name="identifier"]',
          code: 'input[name="code"]',
          submit: 'button[name="method"][value="code"]',
          resend: 'button[name="resend"]',
        },
      }

      before(() => {
        cy.deleteMail()
        cy.useConfigProfile(profile)
        if (app !== "mobile") {
          cy.proxy(app)
        }
        cy.setPostCodeRegistrationHooks([])
        cy.setupHooks("login", "after", "code", [])
      })

      beforeEach(() => {
        const email = gen.email()
        cy.wrap(email).as("email")
        cy.registerWithCode({ email, traits: { "traits.tos": 1 } })

        cy.deleteMail()
        cy.clearAllCookies()
        cy.visit(route)
      })

      it("should be able to sign in with code", () => {
        cy.get("@email").then((email) => {
          cy.get(Selectors[app]["identity"]).clear().type(email.toString())
          cy.submitCodeForm(app)

          cy.getLoginCodeFromEmail(email.toString()).then((code) => {
            cy.get(Selectors[app]["code"]).type(code)

            cy.get(Selectors[app]["submit"]).click()
          })

          cy.location("pathname").should("not.contain", "login")

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
            expect(identity.verifiable_addresses).to.have.length(1)
            expect(identity.verifiable_addresses[0].status).to.equal(
              "completed",
            )
            expect(identity.traits.email).to.equal(email)
          })
        })
      })

      it("should be able to sign in with code on account registered with password", () => {
        const email = gen.email()
        // register account with password
        cy.register({
          email,
          password: gen.password(),
          fields: { "traits.tos": 1 },
        })

        cy.getVerificationCodeFromEmail(email).then((code) => {
          expect(code).to.not.be.empty
          cy.deleteMail()
        })

        cy.clearAllCookies()

        cy.visit(route)

        cy.get(Selectors[app]["identity"]).clear().type(email)
        cy.submitCodeForm(app)

        cy.getLoginCodeFromEmail(email).then((code) => {
          cy.get(Selectors[app]["code"]).type(code)

          cy.get(Selectors[app]["submit"]).click()
        })

        if (app === "express") {
          cy.url().should("match", /\/welcome/)
        } else {
          cy.get('[data-testid="session-content"]').should("contain", email)
        }

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
          expect(identity.verifiable_addresses).to.have.length(1)
          expect(identity.verifiable_addresses[0].status).to.equal("completed")
          expect(identity.traits.email).to.equal(email)
        })
      })

      it("should be able to resend login code", () => {
        cy.get("@email").then((email) => {
          cy.get(Selectors[app]["identity"]).clear().type(email.toString())
          cy.submitCodeForm(app)

          cy.getLoginCodeFromEmail(email.toString()).then((code) => {
            cy.wrap(code).as("code1")
          })

          cy.get(Selectors[app]["resend"]).click()

          cy.getLoginCodeFromEmail(email.toString()).then((code) => {
            cy.wrap(code).as("code2")
          })

          cy.get("@code1").then((code1) => {
            cy.get("@code2").then((code2) => {
              expect(code1).to.not.equal(code2)
            })
          })

          // attempt to submit code 1
          cy.get("@code1").then((code1) => {
            cy.get(Selectors[app]["code"]).clear().type(code1.toString())
          })

          cy.get(Selectors[app]["submit"]).click()

          cy.get("[data-testid='ui/message/4010008']").contains(
            "The login code is invalid or has already been used",
          )

          // attempt to submit code 2
          cy.get("@code2").then((code2) => {
            cy.get(Selectors[app]["code"]).clear().type(code2.toString())
          })

          cy.get(Selectors[app]["submit"]).click()

          if (app === "express") {
            cy.url().should("match", /\/welcome/)
          } else {
            cy.get('[data-testid="session-content"]').should("contain", email)
          }

          if (app === "express") {
            cy.get('a[href*="sessions"').click()
          }

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

          cy.get<Session>("@session").then((session) => {
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

        cy.get(Selectors[app]["identity"]).clear().type(email2)
        cy.submitCodeForm(app)

        cy.getLoginCodeFromEmail(email2).then((code) => {
          cy.get(Selectors[app]["code"]).type(code)
          cy.get(Selectors[app]["submit"]).click()
        })

        if (app === "express") {
          cy.url().should("match", /\/welcome/)
        } else {
          cy.get('[data-testid="session-content"]').should("contain", email)
        }

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
          cy.get('[data-testid="session-content"]').should("contain", email2)
        } else {
          cy.getSession({ expectAal: "aal1", expectMethods: ["code"] }).then(
            (session) => {
              cy.wrap(session).as("session")
            },
          )
        }

        cy.get<Session>("@session").then((session) => {
          expect(session?.identity?.verifiable_addresses).to.have.length(2)
          expect(session?.identity?.verifiable_addresses[0].status).to.equal(
            "completed",
          )
          expect(session.identity.verifiable_addresses[1].status).to.equal(
            "completed",
          )
        })
      })
    })
  })
})

// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { MOBILE_URL, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Login error messages with code method", () => {
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
        },
        express: {
          identity: '[data-testid="login-flow-code"] input[name="identifier"]',
          code: 'input[name="code"]',
        },
        react: {
          identity: 'input[name="identifier"]',
          code: 'input[name="code"]',
        },
      }

      before(() => {
        if (app !== "mobile") {
          cy.proxy(app)
        }
      })

      beforeEach(() => {
        cy.useConfigProfile(profile)
        cy.deleteMail()
        cy.clearAllCookies()

        const email = gen.email()
        cy.wrap(email).as("email")
        cy.registerWithCode({ email, traits: { "traits.tos": 1 } })

        cy.deleteMail()
        cy.clearAllCookies()
        cy.visit(route)
      })

      it("should show error message when account identifier does not exist", () => {
        const email = gen.email()

        cy.get(Selectors[app]["identity"]).type(email)
        cy.submitCodeForm(app)

        cy.get('[data-testid="ui/message/4000035"]').should(
          "contain",
          "This account does not exist or has not setup sign in with code.",
        )
      })

      it("should show error message when code is invalid", () => {
        cy.get("@email").then((email) => {
          cy.get(Selectors[app]["identity"]).clear().type(email.toString())
        })

        cy.submitCodeForm(app)

        cy.get('[data-testid="ui/message/1010014"]').should("exist")

        cy.get(Selectors[app]["code"]).type("123456")
        cy.submitCodeForm(app)

        cy.get('[data-testid="ui/message/4010008"]').should(
          "contain",
          "The login code is invalid or has already been used. Please try again.",
        )
      })

      it("should show error message when identifier has changed", () => {
        cy.get("@email").then((email) => {
          cy.get(Selectors[app]["identity"]).type(email.toString())
        })

        cy.submitCodeForm(app)

        if (app !== "express") {
          cy.intercept("POST", "/self-service/login*", (req) => {
            req.body = {
              ...req.body,
              identifier: gen.email(),
            }
            req.continue()
          }).as("login")
        } else {
          cy.get(Selectors[app]["identity"])
            .type("{selectall}{backspace}", { force: true })
            .type(gen.email(), { force: true })
        }

        cy.get(Selectors[app]["code"]).type("123456")

        cy.submitCodeForm(app)
        if (app !== "express") {
          cy.wait("@login")
        }
        cy.get('[data-testid="ui/message/4000035"]').should(
          "contain",
          "This account does not exist or has not setup sign in with code.",
        )
      })

      it("should show error message when required fields are missing", () => {
        cy.get("@email").then((email) => {
          cy.get(Selectors[app]["identity"]).type(email.toString())
        })

        cy.submitCodeForm(app)

        cy.removeAttribute([Selectors[app]["code"]], "required")
        cy.submitCodeForm(app)

        if (app === "mobile") {
          cy.get('[data-testid="field/code"]').should(
            "contain",
            "Property code is missing",
          )
        } else {
          cy.get('[data-testid="ui/message/4000002"]').should(
            "contain",
            "Property code is missing",
          )
        }

        cy.get(Selectors[app]["code"]).type("123456")
        cy.removeAttribute([Selectors[app]["identity"]], "required")

        cy.get(Selectors[app]["identity"]).type("{selectall}{backspace}", {
          force: true,
        })

        cy.submitCodeForm(app)
        if (app === "mobile") {
          cy.get('[data-testid="field/identifier"]').should(
            "contain",
            "Property identifier is missing",
          )
        } else if (app === "react") {
          // The backspace trick is not working in React.
          cy.get('[data-testid="ui/message/4010008"]').should(
            "contain",
            "code is invalid",
          )
        } else {
          cy.get('[data-testid="ui/message/4000002"]').should(
            "contain",
            "Property identifier is missing",
          )
        }
      })

      it("should show error message when code is expired", () => {
        cy.updateConfigFile((config) => {
          config.selfservice.methods.code = {
            passwordless_enabled: true,
            config: {
              lifespan: "1ns",
            },
          }
          return config
        }).then(() => {
          cy.visit(route)
        })

        cy.get("@email").then((email) => {
          cy.get(Selectors[app]["identity"]).type(email.toString())
        })
        cy.submitCodeForm(app)

        cy.get("@email").then((email) => {
          cy.getLoginCodeFromEmail(email.toString()).then((code) => {
            cy.get(Selectors[app]["code"]).type(code)
          })
        })

        cy.submitCodeForm(app)

        // the react app does not show the error message for 410 errors
        // it just creates a new flow
        if (app === "express") {
          cy.get('[data-testid="ui/message/4010001"]').should(
            "contain",
            "The login flow expired",
          )
        } else {
          cy.get(Selectors[app]["identity"]).should("be.visible")
        }

        cy.noSession()
      })
    })
  })
})

// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0
import { UiNode } from "@ory/kratos-client"
import { gen, MOBILE_URL } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Registration error messages with code method", () => {
  ;[
    {
      route: express.registration,
      app: "express" as "express",
      profile: "code",
    },
    {
      route: react.registration,
      app: "react" as "react",
      profile: "code",
    },
    {
      route: MOBILE_URL + "/Registration",
      app: "mobile" as "mobile",
      profile: "code",
    },
  ].forEach(({ route, profile, app }) => {
    describe(`for app ${app}`, () => {
      const Selectors = {
        mobile: {
          identifier: "[data-testid='field/identifier']",
          email: "[data-testid='field/traits.email']",
          tos: "[data-testid='traits.tos']",
          code: "[data-testid='field/code']",
        },
        express: {
          identifier:
            "[data-testid='registration-flow-code'] input[name='identifier']",
          email:
            "[data-testid='registration-flow-code'] input[name='traits.email']",
          tos: "[data-testid='registration-flow-code'] [name='traits.tos'] + label",
          code: "input[name='code']",
        },
        react: {
          identifier: "input[name='identifier']",
          email: "input[name='traits.email']",
          tos: "[name='traits.tos'] + label",
          code: "input[name='code']",
        },
      }

      before(() => {
        if (app !== "mobile") {
          cy.proxy(app)
        }
        cy.useConfigProfile(profile)
        cy.deleteMail()
      })

      beforeEach(() => {
        cy.deleteMail()
        cy.clearAllCookies()
        cy.visit(route)
      })

      it("should show error message when code is invalid", () => {
        const email = gen.email()

        cy.get(Selectors[app]["email"]).type(email)
        cy.get(Selectors[app]["tos"]).click()

        cy.submitCodeForm(app)

        cy.get('[data-testid="ui/message/1040005"]').should("be.visible")

        cy.get(Selectors[app]["code"]).type("123456")
        cy.submitCodeForm(app)

        cy.get('[data-testid="ui/message/4040003"]').should(
          "contain",
          "The registration code is invalid or has already been used. Please try again.",
        )
      })

      it("should show error message when traits have changed", () => {
        const email = gen.email()

        cy.get(Selectors[app]["email"]).type(email)
        cy.get(Selectors[app]["tos"]).click()

        cy.submitCodeForm(app)
        cy.get('[data-testid="ui/message/1040005"]').should("be.visible")

        if (app !== "express") {
          // the mobile app doesn't render hidden fields in the DOM
          // we need to replace the request body
          cy.intercept("POST", "/self-service/registration*", (req) => {
            req.body = {
              ...req.body,
              "traits.email": "changed-email@email.com",
            }
            req.continue()
          }).as("registration")
        } else {
          cy.get(Selectors[app]["email"])
            .type("{selectall}{backspace}", { force: true })
            .type("changed-email@email.com", { force: true })
        }

        cy.get(Selectors[app]["code"]).type("123456")
        cy.submitCodeForm(app)

        if (app !== "express") {
          cy.wait("@registration")
        }

        cy.get('[data-testid="ui/message/4000036"]').should(
          "contain",
          "The provided traits do not match the traits previously associated with this flow.",
        )
      })

      it("should show error message when required fields are missing", () => {
        const email = gen.email()

        cy.get(Selectors[app]["email"]).type(email)
        cy.get(Selectors[app]["tos"]).click()

        cy.submitCodeForm(app)
        cy.get('[data-testid="ui/message/1040005"]').should("be.visible")

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

        if (app !== "express") {
          // the mobile app doesn't render hidden fields in the DOM
          // we need to replace the request body
          cy.intercept("POST", "/self-service/registration*", (req) => {
            delete req.body["traits.email"]
            req.continue((res) => {
              const emailInput = res.body.ui.nodes.find(
                (n: UiNode) =>
                  "name" in n.attributes &&
                  n.attributes.name === "traits.email",
              )
              expect(emailInput).to.not.be.undefined
              expect(emailInput.messages).to.not.be.undefined
              expect(emailInput.messages[0].text).to.contain("email is missing")
            })
          }).as("registration")
        } else {
          cy.get(Selectors[app]["email"]).type("{selectall}{backspace}", {
            force: true,
          })
          cy.removeAttribute([Selectors[app]["email"]], "required")
        }
        cy.get(Selectors[app]["code"]).type("123456")

        cy.submitCodeForm(app)

        if (app !== "express") {
          cy.wait("@registration")
        } else {
          cy.get('[data-testid="ui/message/4000002"]').should(
            "contain",
            "Property email is missing",
          )
        }
      })

      it("should show error message when code is expired", () => {
        cy.updateConfigFile((config) => {
          config.selfservice.methods.code.config.lifespan = "1ns"
          return config
        })
        cy.visit(route)

        const email = gen.email()
        cy.get(Selectors[app]["email"]).type(email)
        cy.get(Selectors[app]["tos"]).click()

        cy.submitCodeForm(app)
        cy.get('[data-testid="ui/message/1040005"]').should("be.visible")

        cy.getRegistrationCodeFromEmail(email).then((code) => {
          cy.get(Selectors[app]["code"]).type(code)
          cy.submitCodeForm(app)
        })

        // in the react spa app we don't show the 410 gone error. we create a new flow.
        if (app === "express") {
          cy.get('[data-testid="ui/message/4040001"]').should(
            "contain",
            "The registration flow expired",
          )
        } else {
          cy.get(Selectors[app]["email"]).should("be.visible")
        }

        cy.noSession()

        cy.updateConfigFile((config) => {
          config.selfservice.methods.code.config.lifespan = "1h"
          return config
        })
      })
    })
  })
})

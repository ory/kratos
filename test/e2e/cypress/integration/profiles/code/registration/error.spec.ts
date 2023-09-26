// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0
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
          email: "[data-testid='traits.email']",
          tos: "[data-testid='traits.tos']",
          code: "[data-testid='field/code']",
        },
        express: {
          identifier: "input[name='identifier']",
          email: "input[name='traits.email']",
          tos: "[name='traits.tos'] + label",
          code: "input[name='code']",
        },
      }
      Selectors["react"] = Selectors["express"]

      before(() => {
        cy.proxy(app)
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

        cy.get('[data-testid="ui/message/1040005"]').should(
          "contain",
          "An email containing a code has been sent to the email address you provided",
        )

        cy.get(Selectors[app]["code"]).type("invalid-code")
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
        cy.get('[data-testid="ui/message/1040005"]').should(
          "contain",
          "An email containing a code has been sent to the email address you provided",
        )

        cy.get(Selectors[app]["email"])
          .type("{selectall}{backspace}", { force: true })
          .type("changed-email@email.com", { force: true })
        cy.get(Selectors[app]["code"]).type("invalid-code")
        cy.submitCodeForm(app)

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
        cy.get('[data-testid="ui/message/1040005"]').should(
          "contain",
          "An email containing a code has been sent to the email address you provided",
        )

        cy.removeAttribute([Selectors[app]["code"]], "required")

        cy.submitCodeForm(app)
        cy.get('[data-testid="ui/message/4000002"]').should(
          "contain",
          "Property code is missing",
        )

        cy.get(Selectors[app]["email"]).clear()
        cy.get(Selectors[app]["code"]).type("invalid-code")
        cy.removeAttribute([Selectors[app]["email"]], "required")

        cy.submitCodeForm(app)
        cy.get('[data-testid="ui/message/4000002"]').should(
          "contain",
          "Property email is missing",
        )
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
        cy.get('[data-testid="ui/message/1040005"]').should(
          "contain",
          "An email containing a code has been sent to the email address you provided",
        )

        cy.getRegistrationCodeFromEmail(email).should((code) => {
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

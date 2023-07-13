// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0
import { appPrefix, APP_URL, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Registration error messages with code method", () => {
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
      })
    })

    beforeEach(() => {
      cy.deleteMail()
      cy.clearAllCookies()
      cy.visit(route)
    })

    it("should show error message when code is invalid", () => {
      const email = gen.email()

      cy.get(
        'form[data-testid="registration-flow-code"] input[name="traits.email"]',
      ).type(email)
      cy.submitCodeForm()

      cy.url().should("contain", "registration")
      cy.get('[data-testid="ui/message/1040005"]').should(
        "contain",
        "An email containing a code has been sent to the email address you provided",
      )

      cy.get(
        'form[data-testid="registration-flow-code"] input[name="code"]',
      ).type("invalid-code")
      cy.submitCodeForm()

      cy.get('[data-testid="ui/message/4040003"]').should(
        "contain",
        "The registration code is invalid or has already been used. Please try again.",
      )
    })

    it("should show error message when traits have changed", () => {
      const email = gen.email()

      cy.get(
        'form[data-testid="registration-flow-code"] input[name="traits.email"]',
      ).type(email)
      cy.submitCodeForm()

      cy.url().should("contain", "registration")
      cy.get(
        'form[data-testid="registration-flow-code"] input[name="traits.email"]',
      )
        .clear()
        .type("changed-email@email.com")
      cy.get(
        'form[data-testid="registration-flow-code"] input[name="code"]',
      ).type("invalid-code")
      cy.submitCodeForm()

      cy.get('[data-testid="ui/message/4000029"]').should(
        "contain",
        "The provided traits do not match the traits previously associated with this flow.",
      )
    })

    it("should show error message when code is expired", () => {
      cy.updateConfigFile((config) => {
        config.selfservice.methods.code.config.lifespan = "1ns"
        return config
      })

      const email = gen.email()

      cy.get(
        'form[data-testid="registration-flow-code"] input[name="traits.email"]',
      ).type(email)
      cy.submitCodeForm()

      cy.url().should("contain", "registration")
      cy.getRegistrationCodeFromEmail(email).then((code) => {
        cy.get(
          'form[data-testid="registration-flow-code"] input[name="code"]',
        ).type(code)
        cy.submitCodeForm()
      })

      cy.get('[data-testid="ui/message/4040001"]').should(
        "contain",
        "The registration flow expired",
      )

      cy.updateConfigFile((config) => {
        config.selfservice.methods.code.config.lifespan = "1h"
        return config
      })
    })

    it("should show error message when required fields are missing", () => {
      const email = gen.email()

      cy.get(
        'form[data-testid="registration-flow-code"] input[name="traits.email"]',
      ).type(email)
      cy.submitCodeForm()

      cy.url().should("contain", "registration")

      cy.removeAttribute(
        ['form[data-testid="registration-flow-code"] input[name="code"]'],
        "required",
      )
      cy.submitCodeForm()

      cy.get('[data-testid="ui/message/4000002"]').should(
        "contain",
        "Property code is missing",
      )

      cy.get(
        'form[data-testid="registration-flow-code"] input[name="traits.email"]',
      ).clear()
      cy.get(
        'form[data-testid="registration-flow-code"] input[name="code"]',
      ).type("invalid-code")
      cy.removeAttribute(
        [
          'form[data-testid="registration-flow-code"] input[name="traits.email"]',
        ],
        "required",
      )

      cy.submitCodeForm()
      cy.get('[data-testid="ui/message/4000002"]').should(
        "contain",
        "Property email is missing",
      )
    })
  })
})

// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, APP_URL, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Login error messages with code method", () => {
  ;[
    {
      route: express.login,
      app: "express" as "express",
      profile: "code",
    },
    // {
    //   route: react.login,
    //   app: "react" as "react",
    //   profile: "code",
    // },
  ].forEach(({ route, profile, app }) => {
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
      const email = gen.email()
      cy.wrap(email).as("email")
      cy.registerWithCode({ email })

      cy.deleteMail()
      cy.clearAllCookies()
    })

    it("should show error message when account identifier does not exist", () => {
      const email = gen.email()

      cy.get(
        'form[data-testid="login-flow-code"] input[name="identifier"]',
      ).type(email)
      cy.submitCodeForm()

      cy.url().should("contain", "login")

      cy.get('[data-testid="ui/message/4000028"]').should(
        "contain",
        "This account does not exist or has not setup sign in with code.",
      )
    })

    it("should show error message when code is invalid", () => {
      cy.get("@email").then((email) => {
        cy.get('form[data-testid="login-flow-code"] input[name="identifier"]')
          .clear()
          .type(email.toString())
      })

      cy.submitCodeForm()

      cy.url().should("contain", "login")
      cy.get('[data-testid="ui/message/1010014"]').should(
        "contain",
        "An email containing a code has been sent to the email address you provided",
      )

      cy.get('form[data-testid="login-flow-code"] input[name="code"]').type(
        "invalid-code",
      )
      cy.submitCodeForm()

      cy.get('[data-testid="ui/message/4010008"]').should(
        "contain",
        "The login code is invalid or has already been used. Please try again.",
      )
    })

    it("should show error message when identifier has changed", () => {
      cy.get("@email").then((email) => {
        cy.get(
          'form[data-testid="login-flow-code"] input[name="identifier"]',
        ).type(email.toString())
      })

      cy.submitCodeForm()

      cy.url().should("contain", "login")
      cy.get('form[data-testid="login-flow-code"] input[name="identifier"]')
        .clear()
        .type(gen.email())
      cy.get('form[data-testid="login-flow-code"] input[name="code"]').type(
        "invalid-code",
      )
      cy.submitCodeForm()

      cy.get('[data-testid="ui/message/4000028"]').should(
        "contain",
        "This account does not exist or has not setup sign in with code.",
      )
    })

    it("should show error message when code is expired", () => {
      cy.clearAllCookies()

      cy.updateConfigFile((config) => {
        config.selfservice.methods.code.config.lifespan = "1ns"
        return config
      })

      cy.visit(route)

      cy.get("@email").then((email) => {
        cy.get(
          'form[data-testid="login-flow-code"] input[name="identifier"]',
        ).type(email.toString())
      })
      cy.submitCodeForm()

      cy.url().should("contain", "login")

      cy.get("@email").then((email) => {
        cy.getLoginCodeFromEmail(email.toString()).then((code) => {
          cy.get('form[data-testid="login-flow-code"] input[name="code"]').type(
            code,
          )
          cy.submitCodeForm()
        })
      })

      cy.get('[data-testid="ui/message/4040001"]').should(
        "contain",
        "The login flow expired",
      )

      cy.updateConfigFile((config) => {
        config.selfservice.methods.code.config.lifespan = "1h"
        return config
      })
    })

    it("should show error message when required fields are missing", () => {
      const email = gen.email()

      cy.get(
        'form[data-testid="login-flow-code"] input[name="identifier"]',
      ).type(email)
      cy.submitCodeForm()

      cy.url().should("contain", "login")

      cy.removeAttribute(
        ['form[data-testid="login-flow-code"] input[name="code"]'],
        "required",
      )
      cy.submitCodeForm()

      cy.get('[data-testid="ui/message/4000002"]').should(
        "contain",
        "Property code is missing",
      )

      cy.get(
        'form[data-testid="login-flow-code"] input[name="identifier"]',
      ).clear()
      cy.get('form[data-testid="login-flow-code"] input[name="code"]').type(
        "invalid-code",
      )
      cy.removeAttribute(
        ['form[data-testid="login-flow-code"] input[name="identifier"]'],
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

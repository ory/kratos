// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

describe("Basic email profile with failing login flows", () => {
  ;[
    {
      route: express.login,
      app: "express" as "express",
      profile: "email",
    },
    {
      route: react.login,
      app: "react" as "react",
      profile: "spa",
    },
  ].forEach(({ route, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.clearAllCookies()
        cy.visit(route)
      })

      it("fails when CSRF cookies are missing", () => {
        cy.get(`${appPrefix(app)}input[name="identifier"]`).type(
          "i-do-not-exist",
        )
        cy.get('input[name="password"]').type("invalid-password")

        cy.shouldHaveCsrfError({ app })
      })

      it("fails when a disallowed return_to url is requested", () => {
        cy.shouldErrorOnDisallowedReturnTo(
          route + "?return_to=https://not-allowed",
          { app },
        )
      })

      describe("shows validation errors when invalid signup data is used", () => {
        it("should show an error when the identifier is missing", () => {
          // the browser will prevent the form from submitting if the fields are empty since they are required
          // here we just remove the required attribute to make the form submit
          cy.removeAttribute(
            ['input[name="identifier"]', 'input[name="password"]'],
            "required",
          )
          cy.submitPasswordForm()
          cy.get('*[data-testid="ui/message/4000002"]').should(
            "contain.text",
            "Property identifier is missing",
          )
          cy.get('*[data-testid="ui/message/4000002"]').should(
            "contain.text",
            "Property password is missing",
          )
        })

        it("should show an error when the password is missing", () => {
          const identity = gen.email()
          cy.get('input[name="identifier"]')
            .type(identity)
            .should("have.value", identity)

          // the browser will prevent the form from submitting if the fields are empty since they are required
          // here we just remove the required attribute to make the form submit
          cy.removeAttribute(['input[name="password"]'], "required")

          cy.submitPasswordForm()
          cy.get('*[data-testid^="ui/message/"]')
            .invoke("text")
            .then((text) => {
              expect(text.trim()).to.be.oneOf([
                "length must be >= 1, but got 0",
                "Property password is missing.",
              ])
            })
        })

        it("should show fail to sign in", () => {
          cy.get('input[name="identifier"]').type("i-do-not-exist")
          cy.get('input[name="password"]').type("invalid-password")

          cy.submitPasswordForm()
          cy.get('*[data-testid="ui/message/4000006"]').should(
            "contain.text",
            "credentials are invalid",
          )
        })
      })
    })
  })
})

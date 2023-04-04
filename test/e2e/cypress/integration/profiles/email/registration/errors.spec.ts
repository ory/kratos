// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, gen } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

describe("Registration failures with email profile", () => {
  ;[
    {
      route: express.registration,
      app: "express" as "express",
      profile: "email",
    },
    {
      route: react.registration,
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
        cy.visit(route)
      })

      const identity = gen.email()
      const password = gen.password()

      it("fails when CSRF cookies are missing", () => {
        cy.get(`${appPrefix(app)} input[name="traits.website"]`).type(
          "https://www.ory.sh",
        )
        cy.get('input[name="traits.email"]')
          .type(identity)
          .should("have.value", identity)
        cy.get('input[name="password"]')
          .type("12345678")
          .should("have.value", "12345678")

        cy.shouldHaveCsrfError({ app })
      })

      it("fails when a disallowed return_to url is requested", () => {
        cy.shouldErrorOnDisallowedReturnTo(
          route + "?return_to=https://not-allowed",
          { app },
        )
      })

      describe("show errors when invalid signup data is used", () => {
        it("should show an error when the password has leaked before", () => {
          cy.get('input[name="traits.website"]').type("https://www.ory.sh")
          cy.get('input[name="traits.email"]')
            .type(identity)
            .should("have.value", identity)
          cy.get('input[name="password"]')
            .type("12345678")
            .should("have.value", "12345678")

          cy.submitPasswordForm()
          cy.get('[data-testid="ui/message/4000005"]').should(
            "contain.text",
            "data breaches",
          )
        })

        it("should show an error when the password is too similar", () => {
          cy.get('input[name="traits.website"]').type("https://www.ory.sh")
          cy.get('input[name="traits.email"]').type(identity)
          cy.get('input[name="password"]').type(identity)

          cy.submitPasswordForm()
          cy.get('[data-testid="ui/message/4000005"]').should(
            "contain.text",
            "too similar",
          )
        })

        it("should show an error when the password is empty", () => {
          cy.get('input[name="traits.website"]').type("https://www.ory.sh")
          cy.get('input[name="traits.email"]').type(identity)

          // the browser will prevent the form from being submitted if the input field is required
          // we should remove the required attribute to simulate the data not being sent
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

        it("should show an error when the email is empty", () => {
          cy.get('input[name="traits.website"]').type("https://www.ory.sh")
          cy.get('input[name="password"]').type(password)

          // the browser will prevent the form from being submitted if the input field is required
          // we should remove the required attribute to simulate the data not being sent
          cy.removeAttribute(['input[name="traits.email"]'], "required")

          cy.submitPasswordForm()
          cy.get('*[data-testid^="ui/message/"]')
            .invoke("text")
            .then((text) => {
              expect(text.trim()).to.be.oneOf([
                '"" is not valid "email"',
                "length must be >= 3, but got 0",
                "Property email is missing.",
              ])
            })
        })

        it("should show an error when the email is not an email", () => {
          cy.get('input[name="traits.website"]').type("https://www.ory.sh")
          cy.get('input[name="password"]').type(password)

          // the browser will prevent the form from being submitted if the input data doesn't conform to the input field type
          // in this case an invalid email will prevent the form from being submitted by the browser
          // we should remove it to ensure kratos is validating the payload
          cy.get('input[name="traits.email"]').then(($el) =>
            $el.removeAttr("type"),
          )
          cy.get('input[name="traits.email"]').type("not-an-email")

          cy.submitPasswordForm()
          cy.get(
            '*[data-testid="ui/message/4000001"], *[data-testid="ui/message/4000002"]',
          ).should("exist")
        })

        it("should show a missing indicator if no fields are set", () => {
          // the browser will prevent the form from being submitted if the input field is required
          // we should remove the required attribute to simulate the data not being sent
          cy.removeAttribute(
            [
              'input[name="traits.email"]',
              'input[name="traits.website"]',
              'input[name="password"]',
            ],
            "required",
          )

          cy.submitPasswordForm()
          cy.get(
            '*[data-testid="ui/message/4000001"], *[data-testid="ui/message/4000002"]',
          ).should("exist")
        })

        it("should show an error when the website is not a valid URI", () => {
          cy.get('input[name="traits.website"]')
            .type("1234")
            .then(($input) => {
              expect(
                ($input[0] as HTMLInputElement).validationMessage,
              ).to.contain("URL")
            })
        })

        it("should show an error when the website is too short", () => {
          // the browser will prevent the form from being submitted if the input field is required
          // we should remove the required attribute to simulate the data not being sent
          cy.removeAttribute(
            ['input[name="traits.email"]', 'input[name="password"]'],
            "required",
          )

          cy.get('input[name="traits.website"]').type("http://s")

          cy.submitPasswordForm()
          cy.get('[data-testid="ui/message/4000003"]').should(
            "contain.text",
            "length must be >= 10",
          )
        })

        it("should show an error when required params are missing", () => {
          // the browser will prevent the form from being submitted if the input field is required
          // we should remove it from the DOM entirely to simulate the data not being sent
          cy.get('input[name="traits.website"]').then(($el) => $el.remove())
          cy.get('input[name="traits.email"]').then(($el) => $el.remove())
          cy.get('input[name="password"]').then(($el) => $el.remove())

          cy.submitPasswordForm()
          cy.get('[data-testid="ui/message/4000002"]').should(
            "contain.text",
            "Property website is missing.",
          )
          cy.get('[data-testid="ui/message/4000002"]').should(
            "contain.text",
            "Property email is missing.",
          )
          cy.get('[data-testid="ui/message/4000002"]').should(
            "contain.text",
            "Property password is missing.",
          )
        })

        it("should show an error when the age is too high", () => {
          // the browser will prevent the form from being submitted if the input field is required
          // we should remove the required attribute to simulate the data not being sent
          cy.removeAttribute(
            [
              'input[name="traits.email"]',
              'input[name="traits.website"]',
              'input[name="password"]',
            ],
            "required",
          )

          cy.get('input[name="traits.age"]').type("600")

          cy.submitPasswordForm()
          cy.get('[data-testid="ui/message/4000020"]').should(
            "contain.text",
            "must be <= 300 but found 600",
          )
        })
      })
    })
  })
})

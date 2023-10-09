// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen, MOBILE_URL, website } from "../../../../helpers"
import { testRegistrationWebhook } from "../../../../helpers/webhook"

context("Mobile Profile", () => {
  describe("Login Flow Success", () => {
    before(() => {
      cy.useConfigProfile("mobile")
    })

    beforeEach(() => {
      cy.deleteMail()
      cy.visit(MOBILE_URL + "/Registration")
    })

    it("should sign up and be logged in", () => {
      const email = gen.email()
      const password = gen.password()

      cy.get('input[data-testid="traits.email"]').type(email)
      cy.get('input[data-testid="password"]').type(password)
      cy.get('input[data-testid="traits.website"]').type(website)
      cy.get('div[data-testid="submit-form"]').click()

      cy.get('[data-testid="session-content"]').should("contain", email)
      cy.get('[data-testid="session-token"]').should("not.be.empty")
    })

    it("should pass transient_payload to webhook", () => {
      testRegistrationWebhook(
        (hooks) => cy.setupHooks("registration", "after", "password", hooks),
        () => {
          const email = gen.email()
          const password = gen.password()

          cy.get('input[data-testid="traits.email"]').type(email)
          cy.get('input[data-testid="password"]').type(password)
          cy.get('input[data-testid="traits.website"]').type(website)
          cy.get('div[data-testid="submit-form"]').click()

          cy.get('[data-testid="session-content"]').should("contain", email)
          cy.get('[data-testid="session-token"]').should("not.be.empty")
        },
      )
    })

    it("should sign up and show verification form", () => {
      cy.enableVerification()
      const email = gen.email()
      const password = gen.password()

      cy.get('input[data-testid="traits.email"]').type(email)
      cy.get('input[data-testid="password"]').type(password)
      cy.get('input[data-testid="traits.website"]').type(website)
      cy.get('div[data-testid="submit-form"]').click()

      cy.get('div[data-testid="field/code"] input').should("be.visible")

      cy.getVerificationCodeFromEmail(email).then((code) => {
        cy.get('div[data-testid="field/code"] input').type(code)
        cy.get(
          'div[data-testid="field/method/code"] div[data-testid=submit-form]',
        ).click()
      })

      cy.get('[data-testid="ui/message/1080002"]').should(
        "have.text",
        "You successfully verified your email address.",
      )

      cy.get("div[data-testid=continue-button]").click()

      cy.get('[data-testid="session-content"]').should("contain", email)
      cy.get('[data-testid="session-token"]').should("not.be.empty")
    })
  })
})

// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen, MOBILE_URL, website } from "../../../../helpers"

context("Mobile Profile", () => {
  describe("Login Flow Success", () => {
    before(() => {
      cy.useConfigProfile("mobile")
    })

    const up = (value) => `not-${value}`

    describe("password", () => {
      const email = gen.email()
      const password = gen.password()

      before(() => {
        cy.registerApi({
          email,
          password,
          fields: { "traits.website": website },
        })
      })

      beforeEach(() => {
        cy.loginMobile({ email, password })
        cy.location("pathname").should("not.contain", "/Login")
        cy.visit(MOBILE_URL + "/Settings")
      })

      it("modifies the password", () => {
        const newPassword = up(password)
        cy.get(
          '*[data-testid="settings-password"] input[data-testid="password"]',
        )
          .clear()
          .type(newPassword)
        cy.get(
          '*[data-testid="settings-password"] div[data-testid="submit-form"]',
        ).click()
        cy.expectSettingsSaved()

        cy.get(
          '*[data-testid="settings-password"] div[data-testid="submit-form"]',
        ).should("not.have.attr", "data-focusable", "false")
        cy.get('*[data-testid="logout"]').click()
        cy.get('input[data-testid="identifier"]').should("exist")

        cy.loginMobile({ email, password })
        cy.get('[data-testid="session-token"]').should("not.exist")
        cy.loginMobile({ email, password: newPassword })
        cy.get('[data-testid="session-token"]').should("not.be.empty")
      })
    })

    describe("profile", () => {
      let email: string
      let password: string

      beforeEach(() => {
        cy.deleteMail()
        email = gen.email()
        password = gen.password()
        cy.registerApi({
          email,
          password,
          fields: { "traits.website": website },
        })
        cy.loginMobile({ email, password })
        cy.location("pathname").should("not.contain", "/Login")
        cy.visit(MOBILE_URL + "/Settings")
      })

      it("modifies an unprotected trait", () => {
        cy.get(
          '*[data-testid="settings-profile"] input[data-testid="traits.website"]',
        )
          .clear()
          .type("https://github.com/ory")
        cy.get(
          '*[data-testid="settings-profile"] div[data-testid="submit-form"]',
        ).click()
        cy.get(
          '*[data-testid="settings-profile"] div[data-testid="submit-form"]',
        ).should("not.have.attr", "data-focusable", "false")

        cy.visit(MOBILE_URL + "/Home")
        cy.get('[data-testid="session-content"]').should(
          "contain",
          "https://github.com/ory",
        )
      })

      it("modifies a protected trait", () => {
        const newEmail = up(email)
        cy.get(
          '*[data-testid="settings-profile"] input[data-testid="traits.email"]',
        )
          .clear()
          .type(newEmail)
        cy.get(
          '*[data-testid="settings-profile"] div[data-testid="submit-form"]',
        ).click()
        cy.get(
          '*[data-testid="settings-profile"] div[data-testid="submit-form"]',
        ).should("not.have.attr", "data-focusable", "false")

        cy.visit(MOBILE_URL + "/Home")
        cy.get('[data-testid="session-content"]').should("contain", newEmail)
      })

      it("shows verification screen after email update", () => {
        cy.enableVerification()
        const newEmail = up(email)
        cy.get(
          '*[data-testid="settings-profile"] input[data-testid="traits.email"]',
        )
          .clear()
          .type(newEmail)
        cy.get(
          '*[data-testid="settings-profile"] div[data-testid="submit-form"]',
        ).click()
        cy.get(
          '*[data-testid="settings-profile"] div[data-testid="submit-form"]',
        ).should("not.have.attr", "data-focusable", "false")

        cy.get('div[data-testid="field/code"] input').should("be.visible")

        cy.getVerificationCodeFromEmail(newEmail).then((code) => {
          cy.get('div[data-testid="field/code"] input').type(code)
          cy.get(
            'div[data-testid="field/method/code"] div[data-testid=submit-form]',
          ).click()
        })

        cy.get('[data-testid="ui/message/1080002"]').should(
          "have.text",
          "You successfully verified your email address.",
        )
      })
    })
  })
})

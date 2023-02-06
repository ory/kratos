// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { APP_URL, gen, MOBILE_URL, website } from "../../../../helpers"
import { authenticator } from "otplib"

context("Mobile Profile", () => {
  describe("TOTP 2FA Flow", () => {
    before(() => {
      cy.useConfigProfile("mobile")
    })

    describe("password", () => {
      let email = gen.email()
      let password = gen.password()

      before(() => {
        cy.clearAllCookies()
      })

      beforeEach(() => {
        email = gen.email()
        password = gen.password()

        cy.registerApi({
          email,
          password,
          fields: { "traits.website": website },
        })
        cy.loginMobile({ email, password })
        cy.visit(MOBILE_URL + "/Settings")
      })

      it("should be able to use both TOTP and lookup", () => {
        // set up totp
        let totpSecret
        cy.get('*[data-testid="field/totp_secret_key/text"]').then(($e) => {
          totpSecret = $e.text().trim()
        })
        cy.get('*[data-testid="field/totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(totpSecret))
        })
        cy.get('*[data-testid="field/method/totp"]').click()
        cy.expectSettingsSaved()

        // Set up backup code
        cy.get('*[data-testid="field/lookup_secret_regenerate/true"]').click()
        let recoveryCodes
        cy.get('*[data-testid="field/lookup_secret_codes/text"]').then(($e) => {
          recoveryCodes = $e.text().trim().split(", ")
        })
        cy.get('*[data-testid="field/lookup_secret_confirm/true"]').click()
        cy.expectSettingsSaved()

        // Lets sign in with TOTP
        cy.visit(MOBILE_URL + "/Login?aal=aal2&refresh=true")
        cy.get('*[data-testid="field/totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(totpSecret))
        })
        cy.get('*[data-testid="field/method/totp"]').click()

        // We have AAL now
        cy.get('[data-testid="session-content"]').should("contain", "aal2")
        cy.get('[data-testid="session-content"]').should("contain", "totp")

        // Lets sign in with lookup secret
        cy.visit(MOBILE_URL + "/Login?aal=aal2&refresh=true")
        cy.get('*[data-testid="field/lookup_secret"]').then(($e) => {
          cy.wrap($e).type(recoveryCodes[0])
        })
        cy.get('*[data-testid="field/method/lookup_secret"]').click()

        // We have AAL now
        cy.get('[data-testid="session-content"]').should("contain", "aal2")
        cy.get('[data-testid="session-content"]').should("contain", "totp")
        cy.get('[data-testid="session-content"]').should(
          "contain",
          "lookup_secret",
        )
      })
    })
  })
})

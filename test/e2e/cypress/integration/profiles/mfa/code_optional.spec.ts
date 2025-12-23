// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen } from "../../../helpers"
import { routes as express } from "../../../helpers/express"

context("2FA code with optional field", () => {
  ;[
    {
      login: express.login,
      settings: express.settings,
      base: express.base,
      app: "express" as "express",
      profile: "mfa-optional",
    },
  ].forEach(({ settings, login, profile, app, base }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      describe("when using highest_available aal with empty optionalMfaEmail field", () => {
        let email: string
        let password: string

        beforeEach(() => {
          email = gen.email()
          password = gen.password()

          // Configure system to use highest available AAL
          cy.useConfig((builder) =>
            builder.longPrivilegedSessionTime().useHighestAvailable(),
          )

          // Register a user without setting up code MFA (empty optionalMfaEmail field)
          cy.register({
            email,
            password,
            fields: { "traits.optionalMfaEmail": "" },
          })
          cy.deleteMail()
        })

        it("should not show 2FA page during login when optionalMfaEmail field is empty", () => {
          // Log out first
          cy.clearAllCookies()

          // Login with the user
          cy.visit(login)
          cy.get('input[name="identifier"]').type(email)
          cy.get('input[name="password"]').type(password)
          cy.get('button[name="method"][value="password"]').click()

          // Should be logged in directly without 2FA page
          cy.getSession({
            expectAal: "aal1",
            expectMethods: ["password"],
          })

          // Verify we're not asked for 2FA when visiting settings
          cy.visit(settings)
          cy.location("pathname").should("contain", "/settings")
          cy.get('input[name="traits.email"]').should("contain.value", email)
        })
      })

      describe("when using highest_available aal with configured optionalMfaEmail field", () => {
        let email: string
        let mfaEmail: string
        let password: string

        beforeEach(() => {
          email = gen.email()
          mfaEmail = gen.email()
          password = gen.password()

          // Configure system to use highest available AAL
          cy.useConfig((builder) =>
            builder.longPrivilegedSessionTime().useHighestAvailable(),
          )

          // Register a user with optionalMfaEmail field set
          cy.register({
            email,
            password,
            fields: { "traits.optionalMfaEmail": mfaEmail },
          })
          cy.deleteMail()
        })

        it("should show 2FA page during login when optionalMfaEmail field is configured", () => {
          // Log out first
          cy.clearAllCookies()

          // Login with the user
          cy.visit(login)
          cy.get('input[name="identifier"]').type(email)
          cy.get('input[name="password"]').type(password)
          cy.get('button[name="method"][value="password"]').click()

          // Should see the code input field
          cy.location("pathname").should("contain", "/login")
          cy.get("input[name='code']").should("be.visible")

          // Get the code from email and enter it
          cy.getLoginCodeFromEmail(mfaEmail).then((code) => {
            cy.get("input[name='code']").type(code)
            cy.contains("Continue").click()
          })

          // Should be logged in with AAL2
          cy.getSession({
            expectAal: "aal2",
            expectMethods: ["password", "code"],
          })
        })

        it("should require 2FA for settings flow when optionalMfaEmail field is configured", () => {
          // Log out first
          cy.clearAllCookies()

          // Try to access settings directly
          cy.visit(settings)

          // Should be redirected to login
          cy.location("pathname").should("contain", "/login")

          // Login with password
          cy.get('input[name="identifier"]').type(email)
          cy.get('input[name="password"]').type(password)
          cy.get('button[name="method"][value="password"]').click()

          // Should see the code input field
          cy.location("pathname").should("contain", "/login")
          cy.get("input[name='code']").should("be.visible")

          // Get the code from email and enter it
          cy.getLoginCodeFromEmail(mfaEmail).then((code) => {
            cy.get("input[name='code']").type(code)
            cy.contains("Continue").click()
          })

          // Go the the settings page again
          cy.visit(settings)

          // Should now be at settings with AAL2
          cy.location("pathname").should("contain", "/settings")
          cy.getSession({
            expectAal: "aal2",
            expectMethods: ["password", "code"],
          })
        })
      })

      describe("when user with configured optionalMfaEmail logs in with aal1 session", () => {
        let email: string
        let mfaEmail: string
        let password: string

        beforeEach(() => {
          email = gen.email()
          mfaEmail = gen.email()
          password = gen.password()

          // Configure system to allow aal1 login but still require highest_available for settings
          cy.useConfig((builder) =>
            builder
              .longPrivilegedSessionTime()
              .useLaxSessionAal()
              .useHighestSettingsFlowAal(),
          )

          // Register a user with optionalMfaEmail field set
          cy.register({
            email,
            password,
            fields: { "traits.optionalMfaEmail": mfaEmail },
          })
          cy.deleteMail()
        })

        it("should not allow access to settings page with aal1 session", () => {
          // Log out first
          cy.clearAllCookies()

          // Login with just password (aal1)
          cy.visit(login)
          cy.get('input[name="identifier"]').type(email)
          cy.get('input[name="password"]').type(password)
          cy.get('button[name="method"][value="password"]').click()

          // Verify we have an aal1 session
          cy.getSession({
            expectAal: "aal1",
            expectMethods: ["password"],
          })

          // Try to access settings page
          cy.visit(settings)

          // Should be redirected to login for 2FA
          cy.location("pathname").should("contain", "/login")

          // Verify we're being asked for 2FA code input
          cy.get("input[name='code']").should("be.visible")
        })
      })
    })
  })
})

// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen, website } from "../../../helpers"
import { routes as express } from "../../../helpers/express"
import { routes as react } from "../../../helpers/react"

context("2FA code", () => {
  ;[
    // {
    //   login: react.login,
    //   settings: react.settings,
    //   base: react.base,
    //   app: "react" as "react",
    //   profile: "spa",
    // },
    {
      login: express.login,
      settings: express.settings,
      base: express.base,
      app: "express" as "express",
      profile: "mfa",
    },
  ].forEach(({ settings, login, profile, app }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      let email: string
      let password: string

      beforeEach(() => {
        email = gen.email()
        password = gen.password()
        cy.useConfig((builder) =>
          builder
            .longPrivilegedSessionTime()
            .useLaxAal()
            .enableCode()
            .enableCodeMFA(),
        )

        cy.register({
          email,
          password,
          fields: { "traits.website": website },
        })
        cy.deleteMail()
        cy.visit(login + "?aal=aal2&via=email")
      })

      it("should be asked to sign in with 2fa if set up", () => {
        cy.get("input[name='identifier']").type(email)
        cy.contains("Continue with code").click()

        cy.get("input[name='code']").should("be.visible")
        cy.getLoginCodeFromEmail(email).then((code) => {
          cy.get("input[name='code']").type(code)
          cy.contains("Submit").click()
        })

        cy.getSession({
          expectAal: "aal2",
          expectMethods: ["password", "code"],
        })
      })

      it("can't use different email in 2fa request", () => {
        cy.get("input[name='identifier']").type(gen.email())
        cy.contains("Continue with code").click()

        cy.get("*[data-testid='ui/message/4010010']").should("be.visible")
        cy.get("input[name='code']").should("not.exist")
        cy.get("input[name='identifier']").should("be.visible")

        // The current session should be unchanged
        cy.getSession({
          expectAal: "aal1",
          expectMethods: ["password"],
        })
      })

      it("entering wrong code should not invalidate correct codes", () => {
        cy.get("input[name='identifier']").type(email)
        cy.contains("Continue with code").click()

        cy.get("input[name='code']").should("be.visible")
        cy.get("input[name='code']").type("123456")
        cy.contains("Submit").click()
        cy.getLoginCodeFromEmail(email).then((code) => {
          cy.get("input[name='code']").type(code)
          cy.contains("Submit").click()
        })

        cy.getSession({
          expectAal: "aal2",
          expectMethods: ["password", "code"],
        })
      })
    })
  })
})

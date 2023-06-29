// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, gen, website } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Testing logout flows", () => {
  ;[
    {
      route: express.login,
      app: "express" as "express",
      profile: "email",
      settings: express.settings,
    },
    {
      route: react.login,
      app: "react" as "react",
      profile: "spa",
      settings: react.settings,
    },
  ].forEach(({ route, profile, app, settings }) => {
    describe(`for app ${app}`, () => {
      let email: string
      let password: string

      before(() => {
        cy.proxy(app)

        email = gen.email()
        password = gen.password()

        cy.useConfigProfile(profile)
        cy.registerApi({
          email,
          password,
          fields: { "traits.website": website },
        })
      })

      beforeEach(() => {
        cy.clearAllCookies()
        cy.login({ email, password, cookieUrl: route })
        cy.visit(route)
      })

      it("should sign out and be able to sign in again", () => {
        cy.getSession()
        cy.getCookie("ory_kratos_session").should("not.be.null")
        if (app === "express") {
          cy.get(
            `${appPrefix(app)} [data-testid="logout"] a:not(.disabled)`,
          ).click()
        } else {
          cy.get(
            `${appPrefix(app)} [data-testid="logout"]:not(.disabled)`,
          ).click()
        }
        cy.getCookie("ory_kratos_session").should("be.null")
        cy.noSession()
        cy.url().should("include", "/login")
      })

      it("should be able to sign out at 2fa page", () => {
        if (app === "react") {
          return
        }
        cy.useLookupSecrets(true)
        cy.sessionRequires2fa()
        cy.getSession({ expectAal: "aal1" })
        cy.getCookie("ory_kratos_session").should("not.be.null")

        // add 2fa to account
        cy.visit(settings)
        cy.get(
          appPrefix(app) + 'button[name="lookup_secret_regenerate"]',
        ).click()
        cy.get('button[name="lookup_secret_confirm"]').click()
        cy.expectSettingsSaved()

        cy.logout()
        cy.visit(route + "?return_to=https://www.ory.sh")

        cy.get('[name="identifier"]').clear().type(email)

        cy.reauth({
          expect: { email, success: false },
          type: { password: password },
        })

        cy.get("a[href*='logout']").click()

        cy.location("host").should("eq", "www.ory.sh")
        cy.useLookupSecrets(false)
      })
    })
  })
})

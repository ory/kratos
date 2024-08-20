// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0
import { appPrefix, gen, website } from "../../../../helpers"
import { routes as express } from "../../../../helpers/express"
import { routes as react } from "../../../../helpers/react"

context("Social Sign In Successes", () => {
  ;[
    {
      login: react.login,
      registration: react.registration,
      settings: react.settings,
      app: "react" as "react",
      profile: "spa",
    },
    {
      login: express.login,
      registration: express.registration,
      settings: express.settings,
      app: "express" as "express",
      profile: "oidc",
    },
  ].forEach(({ login, registration, profile, app, settings }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      beforeEach(() => {
        cy.clearAllCookies()
      })

      it("should be able to sign up, sign out, and then sign in", () => {
        const email = gen.email()
        cy.registerOidc({ app, email, website, route: registration })
        cy.logout()
        cy.noSession()
        cy.loginOidc({ app, url: login })
      })

      it("should be able to sign up and link existing account", () => {
        const email = gen.email()
        const password = gen.password()

        // Create a new account
        cy.registerApi({
          email,
          password,
          fields: { "traits.website": website },
        })

        // Try to log in with the same identifier through OIDC. This should fail and create a new login flow.
        cy.registerOidc({
          app,
          email,
          website,
          expectSession: false,
        })
        cy.noSession()

        // Log in with the same identifier through the login flow. This should link the accounts.
        cy.get('input[name="password"]').type(password)
        cy.submitPasswordForm()
        cy.location("pathname").should("not.contain", "/login")
        cy.getSession()

        // Hydra OIDC should now be linked
        cy.visit(settings)
        cy.get('[value="hydra"]')
          .should("have.attr", "name", "unlink")
          .should("contain.text", "Unlink Ory")
      })

      it("should be able to sign up with redirects", () => {
        const email = gen.email()
        cy.registerOidc({
          app,
          email,
          website,
          route: registration + "?return_to=https://www.example.org/",
        })
        cy.location("href").should("eq", "https://www.example.org/")
        cy.logout()
        cy.noSession()
        cy.loginOidc({
          app,
          url: login + "?return_to=https://www.example.org/",
        })
        cy.location("href").should("eq", "https://www.example.org/")
      })

      it("should be able to log in with upstream parameters", () => {
        const email = gen.email()
        cy.registerOidc({
          app,
          email,
          website,
          route: registration + "?return_to=https://www.example.org/",
        })

        cy.location("href").should("eq", "https://www.example.org/")
        cy.logout()
        cy.noSession()
        cy.intercept("GET", "**/oauth2/auth*").as("getHydraLogin")
        cy.loginOidc({
          app,
          url: login + "?return_to=https://www.example.org/",
          preTriggerHook: () => {
            // add login_hint to upstream_parameters
            // this injects an input element into the form
            cy.addInputElement("form", "upstream_parameters.login_hint", email)
          },
        })
        // once a request to getHydraLogin responds, 'cy.wait' will resolve
        cy.wait("@getHydraLogin")
          .its("request.url")
          .should("include", "login_hint=" + encodeURIComponent(email))

        cy.location("href").should("eq", "https://www.example.org/")
      })
    })
  })
})

// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen, website } from "../../../../helpers"
import { routes as react } from "../../../../helpers/react"
import { routes as express } from "../../../../helpers/express"

context("Social Sign In Successes", () => {
  ;[
    {
      login: react.login,
      registration: react.registration,
      app: "react" as "react",
      profile: "spa",
    },
    {
      login: express.login,
      registration: express.registration,
      app: "express" as "express",
      profile: "oidc",
    },
  ].forEach(({ login, registration, profile, app }) => {
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
    })
  })
})

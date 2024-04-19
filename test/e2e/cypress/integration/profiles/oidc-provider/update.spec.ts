// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { appPrefix, gen, website } from "../../../helpers"
import { routes as express } from "../../../helpers/express"
import { routes as react } from "../../../helpers/react"

context("OpenID Provider Update", () => {
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
        cy.visit(registration)
        cy.setIdentitySchema(
          "file://test/e2e/profiles/oidc/identity.traits.schema.json",
        )
      })

      const shouldSession = (email) => (session) => {
        const { identity } = session
        expect(identity.id).to.not.be.empty
        expect(identity.traits.website).to.equal(website)
        expect(identity.traits.email).to.equal(email)
      }

      it("should be able to sign up, sign out, sign in and then check token", () => {
        const email = gen.email()

        // sign up
        cy.registerOidc({
          app,
          email,
          expectSession: false,
          route: registration,
        })

        cy.get("#registration-password").should("not.exist")
        cy.get(appPrefix(app) + '[name="traits.email"]').should(
          "have.value",
          email,
        )
        cy.get('[data-testid="ui/message/4000002"]').should(
          "contain.text",
          "Property website is missing",
        )

        cy.get('[name="traits.consent"][type="checkbox"]')
          .siblings("label")
          .click()
        cy.get('[name="traits.newsletter"][type="checkbox"]')
          .siblings("label")
          .click()
        cy.get('[name="traits.website"]').type("http://s")

        cy.get('[name="provider"]')
          .should("have.length", 1)
          .should("have.value", "hydra")
          .should("contain.text", "Continue")
          .click()

        cy.get("#registration-password").should("not.exist")
        cy.get('[name="traits.email"]').should("have.value", email)
        cy.get('[name="traits.website"]').should("have.value", "http://s")
        cy.get('[data-testid="ui/message/4000003"]').should(
          "contain.text",
          "length must be >= 10",
        )
        cy.get('[name="traits.website"]')
          .should("have.value", "http://s")
          .clear()
          .type(website)

        cy.get('[name="traits.consent"]').should("be.checked")
        cy.get('[name="traits.newsletter"]').should("be.checked")

        cy.triggerOidc(app)

        cy.location("pathname").should((loc) => {
          expect(loc).to.be.oneOf(["/welcome", "/", "/sessions"])
        })

        // sign out

        cy.logout()
        cy.noSession()

        // sign in
        cy.loginOidc({ app })

        cy.location("pathname").should((loc) => {
          expect(loc).to.be.oneOf(["/welcome", "/", "/sessions"])
        })
        cy.getSession().then((session) => {
          shouldSession(email)(session)
          cy.getFullIdentityById({ id: session.identity.id }).then(
            (identity) => {
              expect(
                identity.credentials.oidc.config.providers[0]
                  .initial_access_token,
              ).to.not.be.empty
              expect(
                identity.credentials.oidc.config.providers[0].initial_id_token,
              ).to.not.be.empty
              expect(
                identity.credentials.oidc.config.providers[0]
                  .initial_refresh_token,
              ).to.not.be.empty
              expect(
                identity.credentials.oidc.config.providers[0].last_access_token,
              ).to.not.be.empty
              expect(
                identity.credentials.oidc.config.providers[0].last_id_token,
              ).to.not.be.empty
              expect(
                identity.credentials.oidc.config.providers[0]
                  .last_refresh_token,
              ).to.not.be.empty

              expect(
                identity.credentials.oidc.config.providers[0]
                  .initial_access_token,
              ).to.not.eq(
                identity.credentials.oidc.config.providers[0].last_access_token,
              )
              expect(
                identity.credentials.oidc.config.providers[0].initial_id_token,
              ).to.not.eq(
                identity.credentials.oidc.config.providers[0].last_id_token,
              )
              expect(
                identity.credentials.oidc.config.providers[0]
                  .initial_refresh_token,
              ).to.not.eq(
                identity.credentials.oidc.config.providers[0]
                  .last_refresh_token,
              )

              expect(
                identity.credentials.oidc.config.providers[0].provider,
              ).to.eq("hydra")
            },
          )
        })
      })
    })
  })
})

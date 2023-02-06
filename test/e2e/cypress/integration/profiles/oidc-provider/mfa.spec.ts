// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { authenticator } from "otplib"
import { gen } from "../../../helpers"
import { routes as express } from "../../../helpers/express"
import * as oauth2 from "../../../helpers/oauth2"
import * as httpbin from "../../../helpers/httpbin"

context("OIDC Provider 2FA", () => {
  const client = {
    auth_endpoint: "http://localhost:4744/oauth2/auth",
    token_endpoint: "http://localhost:4744/oauth2/token",
    id: Cypress.env("OIDC_DUMMY_CLIENT_ID"),
    secret: Cypress.env("OIDC_DUMMY_CLIENT_SECRET"),
    token_endpoint_auth_method: "client_secret_basic",
    grant_types: ["authorization_code", "refresh_token"],
    response_types: ["code", "id_token"],
    scopes: ["openid", "offline", "email", "website"],
    callbacks: [
      "http://localhost:5555/callback",
      "https://httpbin.org/anything",
    ],
  }

  ;[
    {
      login: express.login,
      settings: express.settings,
      base: express.base,
      profile: "oidc-provider-mfa",
      app: "express" as "express",
    },
  ].forEach(({ settings, login, profile, app, base }) => {
    describe(`for app ${app}`, () => {
      let email = gen.email()
      let password = gen.password()
      let secret

      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)

        email = gen.email()
        password = gen.password()

        cy.register({
          email,
          password,
          fields: { "traits.website": "http://t1.local" },
        })
        cy.visit(settings)

        cy.get('[data-testid="node/text/totp_secret_key/text"]').then(($e) => {
          secret = $e.text().trim()
        })
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[name="method"][value="totp"]').click()
        cy.expectSettingsSaved()
        cy.getSession({
          expectAal: "aal2",
          expectMethods: ["password", "totp"],
        })

        cy.clearAllCookies()
      })

      it("should be be asked to sign in with 2fa if set up", () => {
        let url = oauth2.getDefaultAuthorizeURL(client)

        cy.get("body")
          .then((body$) => {
            // Credits https://github.com/suchipi, https://github.com/cypress-io/cypress/issues/944#issuecomment-444312914
            const appWindow = body$[0].ownerDocument.defaultView
            const appIframe = appWindow.parent.document.querySelector("iframe")

            return new Promise((resolve) => {
              appIframe.onload = () => resolve(undefined)
              appWindow.location.href = url
            })
          })
          .then(() => {
            // kratos login ui
            cy.get("[name=identifier]").type(email)
            cy.get("[name=password]").type(password)
            cy.get("[type=submit]").click()

            cy.get('input[name="totp_code"]').then(($e) => {
              cy.wrap($e).type(authenticator.generate(secret))
            })
            cy.get('*[name="method"][value="totp"]').click()

            // consent ui
            cy.get("#openid").click()
            cy.get("#offline").click()
            cy.get("#accept").click()

            let scope = ["offline", "openid"]
            httpbin.checkToken(client, scope, (token: any) => {
              expect(token).to.have.property("access_token")
              expect(token).to.have.property("id_token")
              expect(token).to.have.property("refresh_token")
              expect(token).to.have.property("token_type")
              expect(token).to.have.property("expires_in")
              expect(token.scope).to.equal("offline openid")
              let idToken = JSON.parse(
                decodeURIComponent(
                  escape(window.atob(token.id_token.split(".")[1])),
                ),
              )
              expect(idToken).to.have.property("amr")
              expect(idToken.amr).to.deep.equal(["password", "totp"])
            })

            // We shouldn't need to authenticate again
            url = oauth2.getDefaultAuthorizeURL(client)

            cy.get("body")
              .then((body$) => {
                // Credits https://github.com/suchipi, https://github.com/cypress-io/cypress/issues/944#issuecomment-444312914
                const appWindow = body$[0].ownerDocument.defaultView
                const appIframe =
                  appWindow.parent.document.querySelector("iframe")

                return new Promise((resolve) => {
                  appIframe.onload = () => resolve(undefined)
                  appWindow.location.href = url
                })
              })
              .then(() => {
                // We get the consent screen instead of login
                cy.get("#openid").click()
                cy.get("#offline").click()
                cy.get("#accept").click()

                httpbin.checkToken(client, scope, (token: any) => {
                  expect(token).to.have.property("access_token")
                  expect(token).to.have.property("id_token")
                  expect(token).to.have.property("refresh_token")
                  expect(token).to.have.property("token_type")
                  expect(token).to.have.property("expires_in")
                  expect(token.scope).to.equal("offline openid")
                  let idToken = JSON.parse(
                    decodeURIComponent(
                      escape(window.atob(token.id_token.split(".")[1])),
                    ),
                  )
                  expect(idToken).to.have.property("amr")
                  expect(idToken.amr).to.deep.equal(["password", "totp"])
                })
              })
          })
      })
    })
  })
})

// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { APP_URL, gen } from "../../../helpers"
import * as oauth2 from "../../../helpers/oauth2"
import * as httpbin from "../../../helpers/httpbin"

context("OpenID Provider", () => {
  before(() => {
    cy.useConfigProfile("oidc-provider")
    cy.proxy("express")
  })
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
      "https://ory-network-httpbin-ijakee5waq-ez.a.run.app/anything",
    ],
  }

  it("registration", () => {
    const url = oauth2.getDefaultAuthorizeURL(client)

    cy.visit(url)
    cy.get("[data-testid=signup-link]").click()

    const email = gen.email()
    const password = gen.password()

    cy.get('[name="traits.email"]').type(email)
    cy.get("[name=password]").type(password)
    cy.get('[name="traits.website"]').type("http://example.com")
    cy.get('input[type=checkbox][name="traits.tos"]').click({ force: true })
    cy.get('[name="traits.age"]').type("199")
    cy.get('input[type=checkbox][name="traits.consent"]').click({ force: true })
    cy.get('input[type=checkbox][name="traits.newsletter"]').click({
      force: true,
    })

    cy.get("[type='submit'][value='password']").click()

    cy.get("#openid").click()
    cy.get("#offline").click()
    cy.get("#accept").click()

    const scope = ["offline", "openid"]
    httpbin.checkToken(client, scope, (token: any) => {
      expect(token).to.have.property("access_token")
      expect(token).to.have.property("id_token")
      expect(token).to.have.property("refresh_token")
      expect(token).to.have.property("token_type")
      expect(token).to.have.property("expires_in")
      expect(token.scope).to.equal("offline openid")
      let idToken = JSON.parse(
        decodeURIComponent(escape(window.atob(token.id_token.split(".")[1]))),
      )
      expect(idToken).to.have.property("amr")
      expect(idToken.amr).to.deep.equal(["password"])
    })
  })

  it("registration with session, skip=false and skip=true", () => {
    const email = gen.email()
    const password = gen.password()

    cy.register({
      email,
      password,
      fields: {
        "traits.website": "https://www.ory.sh",
        "traits.tos": "1",
        "traits.age": 22,
      },
    })

    const url = oauth2.getDefaultAuthorizeURL(client)

    cy.request(url).then((res) => {
      const lastResp = res.allRequestResponses[1]["Request URL"]
      const login_challenge = new URL(lastResp).searchParams.get(
        "login_challenge",
      )
      expect(login_challenge).to.not.be.null
      cy.visit(
        APP_URL +
          "/self-service/registration/browser?login_challenge=" +
          login_challenge,
      )
    })

    cy.url().should("contain", "/login")
    cy.get("[data-testid='login-flow']").should("exist")
    cy.get("[data-testid='login-flow'] [name='password']").type(password)
    cy.get(
      "[data-testid='login-flow'] button[name='method'][value='password']",
    ).click()

    // we want to skip the consent flow here
    // so we ask to remember the user
    cy.get("[name='remember']").click()
    cy.get("#openid").click()
    cy.get("#offline").click()
    cy.get("#accept").click()

    const scope = ["offline", "openid"]
    httpbin.checkToken(client, scope, (token: any) => {
      expect(token).to.have.property("access_token")
      expect(token).to.have.property("id_token")
      expect(token).to.have.property("refresh_token")
      expect(token).to.have.property("token_type")
      expect(token).to.have.property("expires_in")
      expect(token.scope).to.equal("offline openid")
      let idToken = JSON.parse(
        decodeURIComponent(escape(window.atob(token.id_token.split(".")[1]))),
      )
      expect(idToken).to.have.property("amr")
      expect(idToken.amr).to.deep.equal(["password", "password"])
    })

    // use the hydra origin to make a new OAuth request from it
    cy.get("body")
      .then((body$) => {
        // Credits https://github.com/suchipi, https://github.com/cypress-io/cypress/issues/944#issuecomment-444312914
        const appWindow = body$[0].ownerDocument.defaultView
        const appIframe = appWindow.parent.document.querySelector("iframe")

        return new Promise((resolve) => {
          appIframe.onload = () => resolve(undefined)
          appWindow.location.href = "http://localhost:4744/health/ready"
        })
      })
      .then(() => {
        // we don't want to redirect here since we only want the login challenge from hydra
        // we reusing the challenge to navigate to the registration page
        cy.request({
          url: oauth2.getDefaultAuthorizeURL(client),
          followRedirect: false,
        })
          .then((res) => {
            expect(res.redirectedToUrl).to.include("login_challenge")
            return new URL(res.redirectedToUrl).searchParams.get(
              "login_challenge",
            )
          })
          .then((login_challenge) => {
            cy.get("body").then((body$) => {
              const appWindow = body$[0].ownerDocument.defaultView
              const appIframe =
                appWindow.parent.document.querySelector("iframe")

              return new Promise((resolve) => {
                appIframe.onload = () => resolve(undefined)
                appWindow.location.href =
                  APP_URL +
                  "/self-service/registration/browser?login_challenge=" +
                  login_challenge
              })
            })
          })
      })

    httpbin.checkToken(client, scope, (token: any) => {
      expect(token).to.have.property("access_token")
      expect(token).to.have.property("id_token")
      expect(token).to.have.property("refresh_token")
      expect(token).to.have.property("token_type")
      expect(token).to.have.property("expires_in")
      expect(token.scope).to.equal("offline openid")
      let idToken = JSON.parse(
        decodeURIComponent(escape(window.atob(token.id_token.split(".")[1]))),
      )
      expect(idToken).to.have.property("amr")
      expect(idToken.amr).to.deep.equal(["password", "password"])
    })
  })
})

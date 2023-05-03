// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen } from "../../../helpers"
import * as uuid from "uuid"
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
      "https://httpbin.org/anything",
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
})

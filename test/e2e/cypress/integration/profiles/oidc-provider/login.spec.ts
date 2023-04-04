// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { gen } from "../../../helpers"
import * as httpbin from "../../../helpers/httpbin"
import * as oauth2 from "../../../helpers/oauth2"

context("OpenID Provider", () => {
  before(() => {
    cy.deleteMail()
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

  it("login", () => {
    const email = gen.email()
    const password = gen.password()
    cy.registerApi({
      email: email,
      password: password,
      fields: { "traits.website": "http://t1.local" },
    })

    const url = oauth2.getDefaultAuthorizeURL(client)

    cy.visit(url)

    // kratos login ui
    cy.get("[name=identifier]").type(email)
    cy.get("[name=password]").type(password)
    cy.get("[type=submit]").click()

    // consent ui
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

  it("login-without-scopes", () => {
    const email = gen.email()
    const password = gen.password()
    cy.registerApi({
      email: email,
      password: password,
      fields: { "traits.website": "http://t1.local" },
    })

    const url = oauth2.getDefaultAuthorizeURL(client)
    cy.visit(url)

    // kratos login ui
    cy.get("[name=identifier]").type(email)
    cy.get("[name=password]").type(password)
    cy.get("[type=submit]").click()

    // consent ui
    cy.get("#accept").click()

    const scope = ["offline", "openid"]
    httpbin.checkToken(client, scope, (token: any) => {
      expect(token).to.have.property("access_token")
      expect(token).not.to.have.property("id_token")
      expect(token).not.to.have.property("refresh_token")
      expect(token).to.have.property("token_type")
      expect(token).to.have.property("expires_in")
      expect(token.scope).to.equal("")
    })
  })

  it("respects-login-remember-config", () => {
    let odicLogin = () => {
      const email = gen.email()
      const password = gen.password()
      cy.registerApi({
        email: email,
        password: password,
        fields: { "traits.website": "http://t1.local" },
      })

      let url = oauth2.getDefaultAuthorizeURL(client)
      cy.visit(url)

      // kratos login ui
      cy.get("[name=identifier]").type(email)
      cy.get("[name=password]").type(password)
      cy.get("[type=submit]").click()

      // consent ui
      cy.get("#accept").click()
    }

    cy.clearAllCookies()
    cy.updateConfigFile((config) => {
      config.session.cookie = config.session.cookie || {}
      config.session.cookie.persistent = true
      config.session.lifespan = "1234s"
      return config
    })

    odicLogin()
    console.log(cy.getCookies())
    cy.getCookie("ory_hydra_session_dev").should("not.be.null")
    cy.getCookie("ory_hydra_session_dev").then((cookie) => {
      let expected = Date.now() / 1000 + 1234
      let precision = 10
      expect(cookie.expiry).to.be.lessThan(expected + precision)
      expect(cookie.expiry).to.be.greaterThan(expected - precision)
    })

    cy.clearAllCookies()
    cy.updateConfigFile((config) => {
      config.session.cookie = config.session.cookie || {}
      config.session.cookie.persistent = false
      return config
    })

    odicLogin()
    cy.getCookie("ory_hydra_session_dev").should("be.null")
  })
})

context("OpenID Provider - change between flows", () => {
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

  function doConsent() {
    cy.url().should("contain", "/consent")

    // consent ui
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
  }

  before(() => {
    cy.deleteMail()
    cy.useConfigProfile("oidc-provider")
    cy.proxy("express")
  })

  it("switch to registration flow", () => {
    const identity = gen.identityWithWebsite()

    const url = oauth2.getDefaultAuthorizeURL(client)

    cy.visit(url)
    cy.get("[href*='/registration']").click()
    cy.url().should("contain", "/registration")

    cy.get("[name='traits.email']").type(identity.email)
    cy.get("[name='password']").type(identity.password)
    cy.get("[name='traits.website']").type(identity.fields["traits.website"])
    cy.get("[type='submit']").click()

    doConsent()
  })

  it("switch to recovery flow", () => {
    cy.deleteMail()
    cy.longRecoveryLifespan()
    cy.longLinkLifespan()
    cy.enableRecovery()
    cy.useRecoveryStrategy("code")

    const identity = gen.identityWithWebsite()
    cy.registerApi(identity)

    const url = oauth2.getDefaultAuthorizeURL(client)
    cy.visit(url)
    cy.get("[href*='/recovery']").click()

    cy.get("input[name='email']").type(identity.email)
    cy.get("button[value='code']").click()
    cy.get('[data-testid="ui/message/1060003"]').should(
      "have.text",
      "An email containing a recovery code has been sent to the email address you provided. If you have not received an email, check the spelling of the address and make sure to use the address you registered with.",
    )

    cy.recoveryEmailWithCode({ expect: { email: identity.email } })
    cy.get("button[value='code']").click()

    cy.get('[data-testid="ui/message/1060001"]', { timeout: 30000 }).should(
      "contain.text",
      "You successfully recovered your account. ",
    )

    cy.getSession()
    cy.location("pathname").should("eq", "/settings")
    // do a password change
    const newPassword = gen.password()
    cy.get('input[name="password"]').clear().type(newPassword)
    cy.get('button[value="password"]').click()

    // we should now end up on the consent screen
    doConsent()
  })
})

import { gen } from "../../../helpers"
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
    id: "dummy-client",
    secret: "secret",
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
    cy.getCookie("oauth2_authentication_session_insecure").should("not.be.null")
    cy.getCookie("oauth2_authentication_session_insecure").then((cookie) => {
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
    cy.getCookie("oauth2_authentication_session_insecure").should("be.null")
  })
})

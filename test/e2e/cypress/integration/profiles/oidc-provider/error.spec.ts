import { routes as express } from "../../../helpers/express"

context("OpenID Provider", () => {
  before(() => {
    cy.useConfigProfile("oidc-provider")
    cy.proxy("express")
  })
  it("should fail with invalid login_challenge", () => {
    cy.visit(express.login + "?login_challenge=not-a-uuid", {
      failOnStatusCode: false,
    }).then((d) => {
      cy.get(`[data-testid="ui/error/message"]`).then((c) => {
        cy.wrap(c[0].textContent).should(
          "contain",
          "the login_challenge parameter is present but invalid or zero UUID",
        )
      })
    })
  })

  it("should fail with zero login_challenge", () => {
    cy.visit(
      express.login + "?login_challenge=00000000-0000-0000-0000-000000000000",
      { failOnStatusCode: false },
    ).then((d) => {
      cy.get(`[data-testid="ui/error/message"]`).then((c) => {
        cy.wrap(c[0].textContent).should(
          "contain",
          "the login_challenge parameter is present but invalid or zero UUID",
        )
      })
    })
  })
})

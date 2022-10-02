import { routes as express } from "../../../helpers/express"

context("OpenID Provider", () => {
  it("should fail with invalid login_challenge", () => {
    cy.visit(express.login + "?login_challenge=not-a-uuid", {
      failOnStatusCode: false,
    }).then((d) => {
      cy.get("code").then((c) => {
        const err = JSON.parse(c[0].textContent)
        cy.wrap(err["code"]).should("equal", 400)
      })
    })
  })

  it("should fail with zero login_challenge", () => {
    cy.visit(
      express.login + "?login_challenge=00000000-0000-0000-0000-000000000000",
      { failOnStatusCode: false },
    ).then((d) => {
      cy.get("code").then((c) => {
        const err = JSON.parse(c[0].textContent)
        cy.wrap(err["code"]).should("equal", 400)
      })
    })
  })
})

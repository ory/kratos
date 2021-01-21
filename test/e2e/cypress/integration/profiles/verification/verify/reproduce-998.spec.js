import {APP_URL, gen} from '../../../../helpers'

context("Verify", () => {
  before(() => {
    cy.deleteMail()
  })

  it("verify with wrong email", () => {
    const identity1 = gen.identity()
    cy.register(identity1)
    cy.deleteMail({atLeast: 1}) // clean up registration email

    cy.login(identity1)
    cy.visit(APP_URL + '/verify')

    cy.get('input[name="email"]').type(identity1.email)
    cy.get('button[type="submit"]').click()

    cy.verifyEmail({ expect: { email: identity1.email } })

    cy.location('pathname').should('eq', '/')

    // up until now copied from success.spec.js
    // identity1 is verified

    cy.logout()

    // registered with other email address
    const identity2 = gen.identity()
    cy.register(identity2)
    cy.deleteMail({ atLeast: 1 }) // clean up registration email

    cy.login(identity2)

    cy.visit(APP_URL + '/verify')

    // request verification link for identity1
    cy.get('input[name="email"]').type(identity1.email)
    cy.get('button[type="submit"]').click()

    cy.verifyEmail({ expect: { email: identity1.email } })

    cy.location('pathname').should('eq', '/')
  })
})

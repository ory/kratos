import { APP_URL, gen, website } from '../../../../helpers'

context('Login', () => {
  beforeEach(() => {
    cy.clearCookies()
    cy.visit(APP_URL + '/auth/login')
  })

  it('should fail when the login request is rejected', () => {
    const email = gen.email()
    cy.get('button[value="hydra"]').click()
    cy.get('#reject').click()
    cy.location('pathname').should('equal', '/auth/login')
    cy.get('.form-errors .message').should(
      'contain.text',
      'login rejected request'
    )
    cy.noSession()
  })

  it('should fail when the consent request is rejected', () => {
    const email = gen.email()
    cy.get('button[value="hydra"]').click()
    cy.get('#username').type(email)
    cy.get('#accept').click()
    cy.get('#reject').click()
    cy.location('pathname').should('equal', '/auth/login')
    cy.get('.form-errors .message').should(
      'contain.text',
      'consent rejected request'
    )
    cy.noSession()
  })

  it('should fail when the id_token is missing', () => {
    const email = gen.email()
    cy.get('button[value="hydra"]').click()
    cy.get('#username').type(email)
    cy.get('#accept').click()
    cy.get('#website').type(website)
    cy.get('#accept').click()
    cy.location('pathname').should('equal', '/auth/login')
    cy.get('.form-errors .message').should('contain.text', 'no id_token')
  })
})

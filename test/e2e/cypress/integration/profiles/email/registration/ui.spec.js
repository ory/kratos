import { APP_URL } from '../../../../helpers'

context('Registration', () => {
  beforeEach(() => {
    cy.visit(APP_URL + '/auth/registration')
  })

  describe('use ui elements', () => {
    it('clicks the visibility toggle to show the password', () => {
      cy.get('input[name="password"]').type('some password')
      cy.get('input[name="password"]').should('have.prop', 'type', 'password')
      cy.get('.password-visibility-toggle').click()
      cy.get('input[name="password"]').should('have.prop', 'type', 'text')
      cy.get('.password-visibility-toggle').click()
      cy.get('input[name="password"]').should('have.prop', 'type', 'password')
    })

    it('clicks the log in link', () => {
      cy.get('a[href*="auth/login"]').click()
      cy.location('pathname').should('include', 'auth/login')
      cy.location('search').should('not.be.empty', 'request')
    })
  })
})
